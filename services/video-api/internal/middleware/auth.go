package middleware
import("net/http";"strings";"github.com/gin-gonic/gin";"github.com/fiap-x/video-platform/video-api/internal/auth")
const SubjectKey="subject"
func Auth(v *auth.Validator) gin.HandlerFunc{return func(c *gin.Context){h:=c.GetHeader("Authorization");if !strings.HasPrefix(h,"Bearer "){c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"code":"unauthorized","message":"Bearer token required"});return};sub,err:=v.Validate(c.Request.Context(),strings.TrimPrefix(h,"Bearer "));if err!=nil{c.AbortWithStatusJSON(http.StatusUnauthorized,gin.H{"code":"unauthorized","message":"Invalid or expired token"});return};c.Set(SubjectKey,sub);c.Next()}}
func Subject(c *gin.Context)string{return c.MustGet(SubjectKey).(string)}
