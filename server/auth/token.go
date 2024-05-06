package auth

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/PaulChristophel/agartha/server/db"
	"gorm.io/gorm"

	model "github.com/PaulChristophel/agartha/server/model/agartha"

	"github.com/PaulChristophel/agartha/server/httputil"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/viper"
)

var jwtSecret = []byte(viper.GetString("secret"))

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Token struct {
	Token string `json:"token" example:"Bearer foo.bar.blah"`
}

// Token generate a new jwt token
//
//	@Description	Creates a new jwt token for the specified user
//	@Tags			auth
//	@Accept			json
//	@Produce		json
//	@Param			credentials	body		Credentials	true	"Credentials"
//	@Success		200			{object}	Token
//	@Failure		400			{object}	httputil.HTTPError400
//	@Failure		401			{object}	httputil.HTTPError401
//	@Failure		403			{object}	httputil.HTTPError403
//	@Failure		500			{object}	httputil.HTTPError500
//	@router			/auth/token [post]
func GetToken(c *gin.Context) {
	var creds Credentials
	db := db.DB // Assuming db.DB is a *gorm.DB instance

	if err := c.ShouldBindJSON(&creds); err != nil {
		httputil.NewError(c, http.StatusBadRequest, "Missing username or password.")
		return
	}

	userData, err := authenticate(creds.Username, creds.Password)
	if err != nil {
		log.Printf("Error authenticating user %s: %+v", creds.Username, err)
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
		log.Printf("Database error: %v", result.Error)
		httputil.NewError(c, http.StatusInternalServerError, "Database error.")
		return
	} else {
		// Optionally update user data or last login time here if necessary
		currentTime := time.Now()     // Get the current time
		user.LastLogin = &currentTime // Update the last login time
		db.Save(&user)
	}

	// Create JWT token
	utime := time.Now().Add(time.Hour * 8).Unix()
	claims := jwt.MapClaims{
		"username":   creds.Username,
		"exp":        utime,
		"expires_at": utime,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		log.Printf("Error signing token: %v", err)
		httputil.NewError(c, http.StatusInternalServerError, "Failed to generate token.")
		return
	}

	// Create session and save session data to session_user_map
	session := sessions.Default(c)
	session.Options(sessions.Options{MaxAge: 28800})
	session.Set("username", creds.Username)
	session.Set("exp", strconv.FormatInt(utime, 10))
	session.Set("expires_at", strconv.FormatInt(utime, 10))
	if err := session.Save(); err != nil {
		log.Printf("Failed to save session: %v", err)
		httputil.NewError(c, http.StatusInternalServerError, "Failed to save session.")
		return
	}

	// Check if a session map already exists for the session ID
	var sessionMap model.SessionUserMap
	findResult := db.Where("session_id = ?", session.ID()).First(&sessionMap)
	if findResult.Error != nil && !errors.Is(findResult.Error, gorm.ErrRecordNotFound) {
		log.Printf("Error retrieving session map: %v", findResult.Error)
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
			log.Printf("Failed to create session map: %v", err)
			httputil.NewError(c, http.StatusInternalServerError, "Failed to save session map.")
			return
		}
	} // No need to update if it exists; if you have fields to update, handle them here

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
