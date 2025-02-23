package core

import (
	"log/slog"

	"github.com/hectorgimenez/d2go/pkg/data"
	"github.com/hectorgimenez/d2go/pkg/data/stat"
	"github.com/hectorgimenez/koolo/internal/context"
)

type BaseCharacter struct {
	*context.Context
}

func (bc BaseCharacter) PreBattleChecks(id data.UnitID, skipOnImmunities []stat.Resist) bool {
	monster, found := bc.Data.Monsters.FindByID(id)
	if !found {
		return false
	}
	for _, i := range skipOnImmunities {
		if monster.IsImmune(i) {
			bc.Logger.Info("Monster is immune! skipping", slog.String("immuneTo", string(i)))
			return false
		}
	}

	return true
}
