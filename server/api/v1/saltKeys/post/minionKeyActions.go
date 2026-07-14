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

// MinionKeyMatch groups minion IDs by Salt key state bucket.
type MinionKeyMatch struct {
	Minions         []string `json:"minions"`
	MinionsDenied   []string `json:"minions_denied"`
	MinionsPre      []string `json:"minions_pre"`
	MinionsRejected []string `json:"minions_rejected"`
}

// MinionKeyActionRequest is the request body for minion key state changes.
type MinionKeyActionRequest struct {
	Match MinionKeyMatch `json:"match"`
}

// MinionKeyActionResponse lists minion IDs affected by a key state change.
type MinionKeyActionResponse struct {
	Minions []string `json:"minions"`
}

// AcceptMinionKeys accepts matched minion keys in the salt_keys table.
//
//	@Summary		Accept minion keys.
//	@Description	Accept matched minion keys in the salt_keys backend. Denied keys are moved into the regular key bank.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	MinionKeyActionResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/minion_keys/accept [post]
//	@Param			req	body	MinionKeyActionRequest	true	"Minion keys to accept."
//	@Security		Bearer
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

// RejectMinionKeys rejects matched minion keys in the salt_keys table.
//
//	@Summary		Reject minion keys.
//	@Description	Reject matched minion keys in the salt_keys backend. Denied keys are moved into the regular key bank as rejected keys.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	MinionKeyActionResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/minion_keys/reject [post]
//	@Param			req	body	MinionKeyActionRequest	true	"Minion keys to reject."
//	@Security		Bearer
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

// DeleteMinionKeys deletes matched minion keys from the salt_keys table.
//
//	@Summary		Delete minion keys.
//	@Description	Delete matched minion keys from the salt_keys backend while preserving Salt key state bucket semantics.
//	@Tags			SaltKeys
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	MinionKeyActionResponse
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		404	{object}	httputil.HTTPError404
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/api/v1/salt_keys/minion_keys/delete [post]
//	@Param			req	body	MinionKeyActionRequest	true	"Minion keys to delete."
//	@Security		Bearer
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

// acceptMinionKeys changes matched minion keys to accepted in one transaction.
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

// rejectMinionKeys changes matched minion keys to rejected in one transaction.
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

// deleteMinionKeys deletes matched minion keys from their current key banks.
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

// upsertDeniedMinionKeys moves denied keys into the regular key bank.
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

// deleteDeniedMinionKeys removes minion keys from the denied key bank.
func deleteDeniedMinionKeys(tx *gorm.DB, minions []string) error {
	return tx.Exec(fmt.Sprintf(`
		DELETE FROM %s
		WHERE bank = ? AND psql_key IN ?
	`, quotedTable()), deniedBank, minions).Error
}

// all returns all minion IDs from every match bucket.
func (match MinionKeyMatch) all() []string {
	all := make([]string, 0, len(match.Minions)+len(match.MinionsDenied)+len(match.MinionsPre)+len(match.MinionsRejected))
	all = append(all, match.Minions...)
	all = append(all, match.MinionsDenied...)
	all = append(all, match.MinionsPre...)
	all = append(all, match.MinionsRejected...)
	return all
}

// quotedTable returns the configured table name as a quoted PostgreSQL identifier.
func quotedTable() string {
	return pq.QuoteIdentifier(table)
}
