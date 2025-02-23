package character

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/skill"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/d2go/pkg/data/state"
	"github.com/hectorgimenez/koolo/internal/action/step"

	"github.com/hectorgimenez/koolo/internal/character/core"
	"github.com/hectorgimenez/koolo/internal/context"
	"github.com/hectorgimenez/koolo/internal/game"
)

type BlizzardSorceress struct {
	core.BaseCharacter
	core.Sorceress
}

func (s BlizzardSorceress) CheckKeyBindings() []skill.ID {

	blizzardSorcRequiredBindings := []skill.ID{skill.Blizzard}

	return s.GetMissingKeyBindings(blizzardSorcRequiredBindings)
}

func (s BlizzardSorceress) KillMonsterSequence(
	monsterSelector func(d game.Data) (data.UnitID, bool),
	skipOnImmunities []stat.Resist,
) error {
	ctx := context.Get()

	completedAttackLoops := 0
	previousUnitID := 0
	previousSelfBlizzard := time.Time{}

	blizzOpts := step.StationaryDistance(
		ctx.CharacterCfg.Character.Sorceress.RightSkillMinDist,
		ctx.CharacterCfg.Character.Sorceress.RightSkillMaxDist,
	)
	lsOpts := step.Distance(
		ctx.CharacterCfg.Character.Sorceress.LeftSkillMinDist,
		ctx.CharacterCfg.Character.Sorceress.LeftSkillMinDist,
	)

	for {
		id, found := monsterSelector(*s.Data)
		if !found {
			return nil
		}
		if previousUnitID != int(id) {
			completedAttackLoops = 0
		}

		if !s.PreBattleChecks(id, skipOnImmunities) {
			return nil
		}

		if completedAttackLoops >= ctx.CharacterCfg.Character.Sorceress.MaxAttacksLoop {
			return nil
		}

		monster, found := s.Data.Monsters.FindByID(id)
		if !found {
			s.Logger.Info("Monster not found", slog.String("monster", fmt.Sprintf("%v", monster)))
			return nil
		}

		// Cast a Blizzard on very close mobs, in order to clear possible trash close the player, every two attack rotations
		if time.Since(previousSelfBlizzard) > time.Second*4 && !s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			for _, m := range s.Data.Monsters.Enemies() {
				if dist := s.PathFinder.DistanceFromMe(m.Position); dist < 4 {
					previousSelfBlizzard = time.Now()
					step.SecondaryAttack(skill.Blizzard, m.UnitID, 1, blizzOpts)
				}
			}
		}

		if s.Data.PlayerUnit.States.HasState(state.Cooldown) {
			step.PrimaryAttack(id, 2, true, lsOpts)
		}

		step.SecondaryAttack(skill.Blizzard, id, 1, blizzOpts)

		completedAttackLoops++
		previousUnitID = int(id)
	}
}
