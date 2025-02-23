package core

import (
	"log/slog"
	"time"

	"github.com/hectorgimenez/koolo/internal/context"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/npc"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/action/step"
	"github.com/hectorgimenez/koolo/internal/game"
)

type Sorceress struct {
	BaseCharacter
}

/*
To leverage this core Sorceress struct, you must embed it in another struct.

That struct should define two methods:

func (s BlizzardSorceress) CheckKeyBindings() []skill.ID {

	blizzardSorcRequiredBindings := []skill.ID{skill.YourRequiredSkill}

	return s.GetMissingKeyBindings(blizzardSorcRequiredBindings)
}

func (s YourCharacterStructHere) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
}

see blizzard_sorceress.go for a reference implementation
*/

func (s Sorceress) PreCTABuffSkills() []skill.ID {
	return []skill.ID{}
}

func (s Sorceress) GetMissingKeyBindings(additionalRequiredSkills []skill.ID) []skill.ID {
	ctx := context.Get()
	requireKeybindings := []skill.ID{skill.Teleport, skill.TomeOfTownPortal, skill.ShiverArmor}
	// Append the additional required skills, if present
	requireKeybindings = append(requireKeybindings, additionalRequiredSkills...)

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		requireKeybindings = append(requireKeybindings, skill.StaticField)
	}

	missingKeybindings := []skill.ID{}

	for _, cskill := range requireKeybindings {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(cskill); !found {
			switch cskill {
			// Since we can have one of 3 armors:
			case skill.ShiverArmor:
				_, found1 := s.Data.KeyBindings.KeyBindingForSkill(skill.FrozenArmor)
				_, found2 := s.Data.KeyBindings.KeyBindingForSkill(skill.ChillingArmor)
				if !found1 && !found2 {
					missingKeybindings = append(missingKeybindings, skill.ShiverArmor)
				}
			default:
				missingKeybindings = append(missingKeybindings, cskill)
			}
		}
	}

	if len(missingKeybindings) > 0 {
		s.Logger.Debug("There are missing required key bindings.", slog.Any("Bindings", missingKeybindings))
	}

	return missingKeybindings
}

func (s Sorceress) BuffSkills() []skill.ID {
	skillsList := make([]skill.ID, 0)
	if _, found := s.Data.KeyBindings.KeyBindingForSkill(skill.EnergyShield); found {
		skillsList = append(skillsList, skill.EnergyShield)
	}

	armors := []skill.ID{skill.ChillingArmor, skill.ShiverArmor, skill.FrozenArmor}
	for _, armor := range armors {
		if _, found := s.Data.KeyBindings.KeyBindingForSkill(armor); found {
			skillsList = append(skillsList, armor)
			return skillsList
		}
	}

	return skillsList
}

func (s Sorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	panic("This should be implemented")
}

func (s Sorceress) killMonsterByName(id npc.ID, monsterType data.MonsterType, skipOnImmunities []stat.Resist) error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		if m, found := d.Monsters.FindOne(id, monsterType); found {
			return m.UnitID, true
		}

		return 0, false
	}, skipOnImmunities)
}

func (s Sorceress) killMonsterWithStatic(bossID npc.ID, monsterType data.MonsterType) error {
	ctx := context.Get()

	for {
		boss, found := s.Data.Monsters.FindOne(bossID, monsterType)
		if !found || boss.Stats[stat.Life] <= 0 {
			return nil
		}

		bossHPPercent := (float64(boss.Stats[stat.Life]) / float64(boss.Stats[stat.MaxLife])) * 100
		thresholdFloat := float64(ctx.CharacterCfg.Character.Sorceress.BossStaticThreshold)

		// Cast Static Field until boss HP is below threshold
		if bossHPPercent > thresholdFloat {
			staticOpts := []step.AttackOption{
				step.Distance(ctx.CharacterCfg.Character.Sorceress.StaticFieldMinDist, ctx.CharacterCfg.Character.Sorceress.StaticFieldMaxDist),
			}
			err := step.SecondaryAttack(skill.StaticField, boss.UnitID, 1, staticOpts...)
			if err != nil {
				s.Logger.Warn("Failed to cast Static Field", slog.String("error", err.Error()))
			}
			continue
		}

		// Switch to primary skill once boss HP is low enough
		return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
			return boss.UnitID, true
		}, nil)
	}
}

func (s Sorceress) KillCountess() error {
	return s.killMonsterByName(npc.DarkStalker, data.MonsterTypeSuperUnique, nil)
}

func (s Sorceress) KillAndariel() error {
	ctx := context.Get()

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Andariel, data.MonsterTypeUnique)
	}
	return s.killMonsterByName(npc.Andariel, data.MonsterTypeUnique, nil)
}

func (s Sorceress) KillSummoner() error {
	return s.killMonsterByName(npc.Summoner, data.MonsterTypeUnique, nil)
}

func (s Sorceress) KillDuriel() error {
	ctx := context.Get()

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Duriel, data.MonsterTypeUnique)
	}
	return s.killMonsterByName(npc.Duriel, data.MonsterTypeUnique, nil)
}

func (s Sorceress) KillCouncil() error {
	return s.KillMonsterSequence(func(d game.Data) (data.UnitID, bool) {
		// Exclude monsters that are not council members
		var councilMembers []data.Monster
		var coldImmunes []data.Monster
		for _, m := range d.Monsters.Enemies() {
			if m.Name == npc.CouncilMember || m.Name == npc.CouncilMember2 || m.Name == npc.CouncilMember3 {
				if m.IsImmune(stat.ColdImmune) {
					coldImmunes = append(coldImmunes, m)
				} else {
					councilMembers = append(councilMembers, m)
				}
			}
		}

		councilMembers = append(councilMembers, coldImmunes...)

		for _, m := range councilMembers {
			return m.UnitID, true
		}

		return 0, false
	}, nil)
}

func (s Sorceress) KillMephisto() error {
	ctx := context.Get()

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Mephisto, data.MonsterTypeUnique)
	}
	return s.killMonsterByName(npc.Mephisto, data.MonsterTypeUnique, nil)
}

func (s Sorceress) KillIzual() error {
	ctx := context.Get()

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.Izual, data.MonsterTypeUnique)
	}
	return s.killMonsterByName(npc.Izual, data.MonsterTypeUnique, nil)
}

func (s Sorceress) KillDiablo() error {
	ctx := context.Get()
	timeout := time.Second * 20
	startTime := time.Now()
	diabloFound := false

	for {
		if time.Since(startTime) > timeout && !diabloFound {
			s.Logger.Error("Diablo was not found, timeout reached")
			return nil
		}

		diablo, found := s.Data.Monsters.FindOne(npc.Diablo, data.MonsterTypeUnique)
		if !found || diablo.Stats[stat.Life] <= 0 {
			// Already dead
			if diabloFound {
				return nil
			}

			// Keep waiting...
			time.Sleep(200)
			continue
		}

		diabloFound = true
		s.Logger.Info("Diablo detected, attacking")

		if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
			return s.killMonsterWithStatic(npc.Diablo, data.MonsterTypeUnique)
		}
		return s.killMonsterByName(npc.Diablo, data.MonsterTypeUnique, nil)
	}
}

func (s Sorceress) KillPindle() error {
	return s.killMonsterByName(npc.DefiledWarrior, data.MonsterTypeSuperUnique, s.CharacterCfg.Game.Pindleskin.SkipOnImmunities)
}

func (s Sorceress) KillNihlathak() error {
	return s.killMonsterByName(npc.Nihlathak, data.MonsterTypeSuperUnique, nil)
}

func (s Sorceress) KillBaal() error {
	ctx := context.Get()

	if ctx.CharacterCfg.Character.Sorceress.UseStaticField {
		return s.killMonsterWithStatic(npc.BaalCrab, data.MonsterTypeUnique)
	}
	return s.killMonsterByName(npc.BaalCrab, data.MonsterTypeUnique, nil)
}
