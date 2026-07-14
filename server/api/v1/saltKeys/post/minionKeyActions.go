package saltKeys

import (
	"fmt"
	"net/http"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

type MinionKeyMatch struct {
	Minions         []string `json:"minions"`
	MinionsDenied   []string `json:"minions_denied"`
	MinionsPre      []string `json:"minions_pre"`
	MinionsRejected []string `json:"minions_rejected"`
}

type MinionKeyActionRequest struct {
	Match MinionKeyMatch `json:"match"`
}

type MinionKeyActionResponse struct {
	Minions []string `json:"minions"`
}

func AcceptMinionKeys(c *gin.Context) {
	var req MinionKeyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	minions := uniqueStrings(req.Match.all())
	if err := acceptMinionKeys(db.DB.Table(table), minions); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	c.JSON(http.StatusOK, MinionKeyActionResponse{Minions: minions})
}

func RejectMinionKeys(c *gin.Context) {
	var req MinionKeyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	minions := uniqueStrings(req.Match.all())
	if err := rejectMinionKeys(db.DB.Table(table), minions); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	c.JSON(http.StatusOK, MinionKeyActionResponse{Minions: minions})
}

func DeleteMinionKeys(c *gin.Context) {
	var req MinionKeyActionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.NewError(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := deleteMinionKeys(db.DB.Table(table), req.Match); err != nil {
		writeSaltKeysError(c, err)
		return
	}

	c.JSON(http.StatusOK, MinionKeyActionResponse{Minions: uniqueStrings(req.Match.all())})
}

func acceptMinionKeys(dbConn *gorm.DB, minions []string) error {
	if err := ensureSaltKeysTable(dbConn); err != nil {
		return err
	}
	if len(minions) == 0 {
		return nil
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(fmt.Sprintf(`
			UPDATE %s
			SET data = jsonb_set(data, '{state}', to_jsonb('accepted'::text), true)
			WHERE bank = ? AND psql_key IN ?
		`, quotedTable()), keysBank, minions).Error; err != nil {
			return err
		}
		if err := upsertDeniedMinionKeys(tx, "accepted", minions); err != nil {
			return err
		}
		return deleteDeniedMinionKeys(tx, minions)
	})
}

func rejectMinionKeys(dbConn *gorm.DB, minions []string) error {
	if err := ensureSaltKeysTable(dbConn); err != nil {
		return err
	}
	if len(minions) == 0 {
		return nil
	}

	return dbConn.Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec(fmt.Sprintf(`
			UPDATE %s
			SET data = jsonb_set(data, '{state}', to_jsonb('rejected'::text), true)
			WHERE bank = ? AND psql_key IN ?
		`, quotedTable()), keysBank, minions).Error; err != nil {
			return err
		}
		if err := upsertDeniedMinionKeys(tx, "rejected", minions); err != nil {
			return err
		}
		return deleteDeniedMinionKeys(tx, minions)
	})
}

func deleteMinionKeys(dbConn *gorm.DB, match MinionKeyMatch) error {
	if err := ensureSaltKeysTable(dbConn); err != nil {
		return err
	}

	keys := uniqueStrings(
		append(
			append([]string{}, match.Minions...),
			append(match.MinionsPre, match.MinionsRejected...)...,
		),
	)
	denied := uniqueStrings(match.MinionsDenied)

	return dbConn.Transaction(func(tx *gorm.DB) error {
		if len(keys) > 0 {
			if err := tx.Exec(fmt.Sprintf(`
				DELETE FROM %s
				WHERE bank = ? AND psql_key IN ?
			`, quotedTable()), keysBank, keys).Error; err != nil {
				return err
			}
		}
		if len(denied) > 0 {
			if err := deleteDeniedMinionKeys(tx, denied); err != nil {
				return err
			}
		}
		return nil
	})
}

func upsertDeniedMinionKeys(tx *gorm.DB, state string, minions []string) error {
	return tx.Exec(fmt.Sprintf(`
		INSERT INTO %s (bank, psql_key, data)
		SELECT ?, psql_key, jsonb_build_object(
			'state', ?,
			'pub', COALESCE(
				CASE WHEN jsonb_typeof(data) = 'array' THEN data->>0 END,
				CASE WHEN jsonb_typeof(data) = 'string' THEN data#>>'{}' END,
				''
			)
		)
		FROM %s
		WHERE bank = ? AND psql_key IN ?
		ON CONFLICT (bank, psql_key) DO UPDATE
		SET data = EXCLUDED.data
	`, quotedTable(), quotedTable()), keysBank, state, deniedBank, minions).Error
}

func deleteDeniedMinionKeys(tx *gorm.DB, minions []string) error {
	return tx.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE bank = ? AND psql_key IN ?
	`, quotedTable()), deniedBank, minions).Error
}

func (match MinionKeyMatch) all() []string {
	all := make([]string, 0, len(match.Minions)+len(match.MinionsDenied)+len(match.MinionsPre)+len(match.MinionsRejected))
	all = append(all, match.Minions...)
	all = append(all, match.MinionsDenied...)
	all = append(all, match.MinionsPre...)
	all = append(all, match.MinionsRejected...)
	return all
}

func quotedTable() string {
	return pq.QuoteIdentifier(table)
}
