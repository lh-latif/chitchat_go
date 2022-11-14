package webserver

// buat middleware untuk check JWT
import (
	"strings"

	db "chitchat/database"

	fiber "github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v4"
	"github.com/valyala/fasthttp"
	"gorm.io/gorm"
)

var secret []byte = []byte("Hello World From Golang")

func InitiateToken(payload jwt.MapClaims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, payload)
	return token.SignedString(secret)
}

func JwtCheckMiddleware(dbConn *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Saat ini masih seperti ini
		reqHeader := c.GetReqHeaders()
		auth, ok := reqHeader["Authorization"]
		if ok {
			_ = auth
			hasBearer := strings.HasPrefix(auth, "Bearer ")
			if hasBearer {
				split := strings.SplitN(auth, " ", 2)
				size := len(split)
				if size == 2 {
					token := split[1]
					_ = token
					tokenstruct, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
						return secret, nil
					})
					if err != nil {
						return c.Status(500).SendString(err.Error())
					} else {
						claims := tokenstruct.Claims.(jwt.MapClaims)
						username := claims["username"]
						user := db.User{}
						dbConn.Where("username = ?", username).Find(&user)
						if user.Username == "" {
							return c.Status(403).Send([]byte{})
						} else {
							var fc *fasthttp.RequestCtx = c.Context()
							fc.SetUserValue("user", user)
						}

						return c.Next()
					}

				}
			}
			return c.Status(400).SendString("something wrong")
		} else {
			return c.Status(403).SendString("forbidden")
		}
	}
}
