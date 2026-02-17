package model

import (
	"encoding/json"
	"log"
	"time"

	"github.com/PaulChristophel/agartha/server/model/custom"
)

// UserSettings represents the settings and permissions for a user.
type UserSettings struct {
	UserID          uint        `json:"user_id" gorm:"primaryKey"`
	Token           string      `json:"token" gorm:"type:varchar(255);not null"`
	Created         time.Time   `json:"created" gorm:"type:timestamp with time zone;not null"`
	SaltPermissions string      `json:"salt_permissions" gorm:"type:text;not null"`
	Settings        custom.JSON `json:"settings" gorm:"type:jsonb;not null"`
	User            AuthUser    `json:"user" gorm:"foreignKey:UserID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
}

// TableName returns the table name for the UserSettings struct.
func (UserSettings) TableName() string {
	return "user_settings"
}

// Create initializes a new UserSettings instance and sets its fields based on the provided parameters.
func (settings *UserSettings) Create(userID uint, saltPermissions, jsonStrPermissions, jsonStrSettings, token string) error {
	*settings = UserSettings{
		UserID:  userID,
		Created: time.Now(),
		Token:   token,
	}

	if err := settings.SetSettingsFromJSON(jsonStrSettings); err != nil {
		log.Printf("Error during user settings creation: %v", err)
		return err
	}
	if err := settings.SetSaltPermissions(jsonStrPermissions); err != nil {
		log.Printf("Error during user settings creation: %v", err)
		return err
	}

	// Note: Uncomment and use the following raw SQL query if you need to use the crypt function for token encryption
	// sql := `INSERT INTO user_settings (user_id, token, created, salt_permissions, settings)
	//         VALUES (?, crypt(?, gen_salt('bf', 8)), ?, ?, ?)
	//         RETURNING user_id;`
	// err := db.Raw(sql, settings.UserID, "Bearer "+token, settings.Created, settings.SaltPermissions, settings.Settings).Scan(&settings.UserID).Error
	return nil
}

// SetSaltPermissions sets the SaltPermissions field from a JSON string.
func (us *UserSettings) SetSaltPermissions(jsonStr string) error {
	if jsonStr == "" {
		us.SaltPermissions = defaultSaltPermissions
	} else {
		us.SaltPermissions = jsonStr
	}
	return nil
}

// SetSettingsFromJSON initializes the Settings field from a JSON string.
func (us *UserSettings) SetSettingsFromJSON(jsonStr string) error {
	if jsonStr == "" {
		jsonStr = defaultSettings
	}

	var js any
	if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
		return err
	}

	us.Settings = custom.JSON{Data: js}
	return nil
}

var defaultSaltPermissions = `[".*", "@jobs", "@runner", "@wheel"]`

var defaultSettings = `
{
	"Home": {
		"JobsChartCard": {
			"filter": "all",
			"period": 7
		}
	},
	"Layout": {
		"dark": true,
		"mini": false,
		"drawer": true
	},
	"RunCard": {
		"tab": "formatted"
	},
	"UserCard": {
		"table": {
			"sortBy": "username",
			"sortDesc": false,
			"itemsPerPage": 10
		}
	},
	"language": "en",
	"JobsTable": {
		"table": {
			"sortBy": "alter_time",
			"sortDesc": true,
			"itemsPerPage": -1
		}
	},
	"KeysTable": {
		"table": {
			"sortBy": "status",
			"sortDesc": true,
			"itemsPerPage": -1
		}
	},
	"EventsTable": {
		"table": {
			"sortBy": "alter_time",
			"sortDesc": true,
			"itemsPerPage": -1
		}
	},
	"MinionDetail": {
		"InfosCard": {
			"tab": "salt"
		},
		"NetworkCard": {
			"tab": "dns"
		},
		"MinionDetailCard": {
			"tab": "grain"
		}
	},
	"MinionsTable": {
		"table": {
			"sortBy": [
				"fqdn"
			],
			"columns": [
				"minion_id",
				"fqdn",
				"os",
				"oscodename"
			],
			"sortDesc": [
				false
			],
			"itemsPerPage": -1
		}
	},
	"UserSettings": {
		"notifs": {
			"event": false,
			"created": true,
			"returned": true,
			"published": true
		},
		"max_notifs": 15
	},
	"ScheduleTable": {
		"table": {
			"sortBy": [
				"minion"
			],
			"sortDesc": [
				false
			],
			"itemsPerPage": 10
		}
	},
	"ConformityTable": {
		"table": {
			"sortBy": "failed",
			"sortDesc": true,
			"itemsPerPage": -1
		}
	},
	"selected_master": "",
	"JobTemplatesTable": {
		"table": {
			"sortBy": "name",
			"sortDesc": false,
			"itemsPerPage": 10
		}
	}
}`
