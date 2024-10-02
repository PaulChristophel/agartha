package auth

import (
	"encoding/xml"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	"github.com/PaulChristophel/agartha/server/logger"
	"gorm.io/gorm"

	model "github.com/PaulChristophel/agartha/server/model/agartha"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type credentials struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Method   string `json:"method" binding:"required"`
}

type Token struct {
	Token string `json:"token" example:"Bearer foo.bar.blah"`
}

type CASServiceResponse struct {
	XMLName               xml.Name              `xml:"serviceResponse"`
	AuthenticationSuccess AuthenticationSuccess `xml:"authenticationSuccess"`
}

type AuthenticationSuccess struct {
	User string `xml:"user"`
}

type AuthMethods struct {
	AuthMethods []string `json:"auth_methods" example:"local"`
}

// Get auth methods
//
//	@Summary		Gets the list of available auth methods.
//	@Description	Gets the list of available auth methods.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	AuthMethods
//	@Failure		400	{object}	httputil.HTTPError400
//	@Failure		401	{object}	httputil.HTTPError401
//	@Failure		403	{object}	httputil.HTTPError403
//	@Failure		500	{object}	httputil.HTTPError500
//	@router			/auth/method [get]
func GetMethod(c *gin.Context) {
	authMethods := []string{"local"}

	if ldapOptions.Server != "ldap.example.com" {
		authMethods = append(authMethods, "ldap")
	}

	if casOptions.Server != "https://cas.example.com" {
		authMethods = append(authMethods, "cas")
	}

	c.JSON(http.StatusOK, gin.H{
		"auth_methods": authMethods,
	})
}

// Token generate a new jwt token
//
//	@Summary		Creates a new jwt token for the specified user.
//	@Description	Creates a new jwt token for the specified user.
//	@Tags			Auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		credentials	true	"Credentials"
//	@Success		200			{object}	Token
//	@Failure		400			{object}	httputil.HTTPError400
//	@Failure		401			{object}	httputil.HTTPError401
//	@Failure		403			{object}	httputil.HTTPError403
//	@Failure		500			{object}	httputil.HTTPError500
//	@router			/auth/token [post]
func RetrieveToken(c *gin.Context) {
	log := logger.GetLogger()
	sugar := log.Sugar()
	var creds credentials
	db := db.DB // Assuming db.DB is a *gorm.DB instance

	if err := c.ShouldBindJSON(&creds); err != nil {
		httputil.NewError(c, http.StatusBadRequest, "Missing username or password.")
		return
	}

	userData, err := auth(creds, c)
	if err != nil {
		sugar.Errorf("Error authenticating user %s: %+v", creds.Username, err)
		httputil.NewError(c, http.StatusUnauthorized, "Invalid credentials.")
		return
	}

	// Check if user exists and handle accordingly
	var user model.AuthUser
	result := db.Where("username = ?", creds.Username).First(&user)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new user if not exist
		user = model.AuthUser{
			Username:    creds.Username,
			Password:    "", // Store a hashed password or omit if password shouldn't be stored
			FirstName:   userData.FirstName,
			LastName:    userData.LastName,
			Email:       userData.Email,
			IsSuperuser: false,
			IsStaff:     false,
			IsActive:    true,
			DateJoined:  time.Now(),
		}
		db.Create(&user)
	} else if result.Error != nil {
		sugar.Errorf("Database error: %v", result.Error)
		httputil.NewError(c, http.StatusInternalServerError, "Database error.")
		return
	} else {
		// Optionally update user data or last login time here if necessary
		currentTime := time.Now()     // Get the current time
		user.LastLogin = &currentTime // Update the last login time
		db.Save(&user)
	}

	// Create JWT token
	maxAge := (time.Hour * 8)
	utime := time.Now().Add(maxAge).Unix()
	claims := jwt.MapClaims{
		"username": creds.Username,
		"user_id":  user.ID,
		"exp":      utime,
		// "expires_at": utime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		sugar.Errorf("Error signing token: %v", err)
		httputil.NewError(c, http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	// Create session and save session data to session_user_map
	session = sessions.Default(c)
	session.Options(sessions.Options{MaxAge: int(maxAge.Seconds())})
	session.Set("username", creds.Username)
	session.Set("user_id", user.ID)
	session.Set("exp", strconv.FormatInt(utime, 10))
	// session.Set("expires_at", strconv.FormatInt(utime, 10))
	if err := session.Save(); err != nil {
		sugar.Errorf("Failed to save session: %v", err)
		httputil.NewError(c, http.StatusInternalServerError, "Failed to save session.")
		return
	}

	// Check if a session map already exists for the session ID
	var sessionMap model.SessionUserMap
	findResult := db.Where("session_id = ?", session.ID()).First(&sessionMap)
	if findResult.Error != nil && !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
		sugar.Errorf("Error retrieving session map: %v", findResult.Error)
		httputil.NewError(c, http.StatusInternalServerError, "Database error during session map retrieval.")
		return
	}

	if findResult.RowsAffected == 0 { // No existing session map, create a new one
		sessionMap = model.SessionUserMap{
			SessionID: session.ID(),
			UserID:    user.ID,
			CreatedAt: time.Now(),
		}
		if err := db.Create(&sessionMap).Error; err != nil {
			sugar.Errorf("Failed to create session map: %v", err)
			httputil.NewError(c, http.StatusInternalServerError, "Failed to save session map.")
			return
		}
	} // No need to update if it exists; if you have fields to update, handle them here

	var settings model.UserSettings
	result = db.Where("user_id = ?", user.ID).First(&settings)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		// Create new user if not exist
		settings = model.UserSettings{
			UserID:          user.ID,
			Created:         time.Now(),
			SaltPermissions: "['.*', '@jobs', '@runner', '@wheel']",
		}
		err := settings.SetSettingsFromJSON("")
		if err != nil {
			sugar.Errorf("Error setting default user settings: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during creation"})
			return
		}
		// Here we use a raw SQL query to insert with the crypt function
		sql := `INSERT INTO user_settings (user_id, token, created, salt_permissions, settings)
                VALUES (?, crypt(?, gen_salt('bf', 8)), ?, ?, ?)
                RETURNING user_id;`
		err = db.Raw(sql, settings.UserID, "Bearer "+tokenString, settings.Created, settings.SaltPermissions, settings.Settings).Scan(&settings.UserID).Error
		if err != nil {
			sugar.Errorf("Error inserting user settings with hashed token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during creation"})
			return
		}

	} else if result.Error != nil {
		sugar.Errorf("Database error: %v", result.Error)
		httputil.NewError(c, http.StatusInternalServerError, "Database error.")
		return
	} else {
		// Update existing user settings
		// Using a raw SQL to update the token with crypt function
		sql := `UPDATE user_settings SET token = crypt(?, gen_salt('bf', 8)) WHERE user_id = ?`
		err := db.Exec(sql, "Bearer "+tokenString, user.ID).Error
		if err != nil {
			sugar.Errorf("Error updating user settings with hashed token: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error during update"})
			return
		}
	}

	c.JSON(http.StatusOK, Token{Token: "Bearer " + tokenString})
}

// // Authenticate authenticates and authorizes a user against an LDAP server
// func authenticate(username, password string) error {
// 	// Connection to the LDAP server
// 	l, err := ldap.DialURL(viper.GetString("ldap_server"))
// 	if err != nil {
// 		return fmt.Errorf("failed to connect: %w", err)
// 	}
// 	ldapDomain := viper.GetString("ldap_domain_default") // Domain suffix
// 	// Ensure username has the domain suffix
// 	if !strings.HasSuffix(username, ldapDomain) {
// 		username += "@" + ldapDomain
// 	}
// 	defer l.Close()

// 	// Attempt to bind with the given username and password for authentication
// 	err = l.Bind(username, password)
// 	if err != nil {
// 		return fmt.Errorf("failed to bind: %w", err)
// 	}

// 	// Rebind as a service account with permissions to search
// 	serviceUser := viper.GetString("ldap_user")
// 	servicePassword := viper.GetString("ldap_password")
// 	err = l.Bind(serviceUser, servicePassword)
// 	if err != nil {
// 		return fmt.Errorf("failed to bind as service user: %w", err)
// 	}

// 	// Authorization: check that the user satisfies the LDAP filter
// 	filter := fmt.Sprintf(viper.GetString("ldap_filter"), username)
// 	searchRequest := ldap.NewSearchRequest(
// 		viper.GetString("ldap_base_dn"), // The base dn to search
// 		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
// 		filter,
// 		[]string{"dn"}, // A list attributes to retrieve
// 		nil,
// 	)

// 	sr, err := l.Search(searchRequest)
// 	if err != nil {
// 		return fmt.Errorf("failed to execute search: %w", err)
// 	}

// 	if len(sr.Entries) != 1 {
// 		return fmt.Errorf("user does not exist or too many entries returned")
// 	}

// 	return nil
// }
