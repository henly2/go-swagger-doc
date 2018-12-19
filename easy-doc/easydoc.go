package easy_doc

import (
	"net/http"
	"github.com/gin-gonic/gin"
	"fmt"
	"strings"
	"github.com/henly2/go-swagger-doc"
	"github.com/gin-contrib/cors"
)

var (
	_engine *gin.Engine
	_routers map[string]*gin.RouterGroup
)

func InitDoc(localesDir, pacDir string, headers []swagger.SecurityDefinition)  {
	swagger.SetLocalesDir(localesDir)

	_routers = make(map[string]*gin.RouterGroup)

	_engine = gin.Default()
	_engine.Use(cors.New(cors.Config{
		AllowAllOrigins: true,
		AllowMethods:    []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:    []string{"Authorization", "X-Requested-With", "X_Requested_With", "Content-Type", "Access-Token", "Accept-Language"},
		//AllowOrigins:     []string{"*"},
		//AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	return true;//origin == "https://github.com"
		//},
		//MaxAge: 12 * time.Hour,
	}))

	_engine.LoadHTMLGlob(pacDir + "/*.html")

	_engine.GET("/documents/:file", func(ctx *gin.Context) {
		file := ctx.Param("file")

		lang := ctx.Query("lang")
		fmt.Println("lang", "--", lang)
		if lang == "" {
			lang = ctx.GetHeader("Accept-Language")
			fmt.Println("Accept-Language", "--", lang)
		}

		if strings.Index(file, ".html") != -1 {
			scheme := "http://"
			if ctx.Request.TLS != nil {
				scheme = "https://"
			}
			host := scheme + ctx.Request.Host

			ctx.HTML(http.StatusOK, file, gin.H{
				"host": host,
				"lang": lang,
			})
		} else {
			ctx.File(pacDir + "/" + file)
		}
	})
	_engine.Use()

	config := swagger.Config{}
	config.Url = "{{.ApiUrl}}"
	config.BasePath = "/"
	config.Title = "{{.ApiDocTitle}}"
	config.Description = "{{.ApiDocDescription}}"
	config.DocVersion = "{{.ApiDocVersion}}"
	config.Headers = headers
	swagger.InitializeApiRoutes(_engine, &config, docLoader)
}

func RunDoc(port string)  {
	_engine.Run(":" + port)
}

func AddDoc(router, path, method string, sp *swagger.StructParam) {
	var r *gin.RouterGroup
	var ok bool

	r, ok = _routers[router]
	if !ok {
		r = _engine.Group(router, func(ctx *gin.Context) {
		})
	}

	swagger.Swagger2(r,path,method, sp)
}

func AddSection(router, section, content string) {
	var r *gin.RouterGroup
	var ok bool

	r, ok = _routers[router]
	if !ok {
		r = _engine.Group(router, func(ctx *gin.Context) {
		})
	}

	swagger.SwaggerSection(r, "", section, content)
}

func docLoader(key string) ([]byte, error) {
	fmt.Println("key:", key)
	return []byte("what"), nil
}

type SayHelloParam struct {
	From   		string `json:"from" doc:"谁发送的"`
	Content  	string `json:"content" doc:"发送的内容"`
}

type SayHelloResponse struct {
	Err   	 int `json:"err" doc:"错误代码"`
	Content  string `json:"content" doc:"内容"`
}

func startSwagger() {
	router := _engine.Group("/api", func(ctx *gin.Context) {

	})
	router.Use(func(ctx *gin.Context) {
		origin := ctx.Request.Header.Get("origin")
		ctx.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		ctx.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		ctx.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, XMLHttpRequest, "+
			"Accept-Encoding, X-CSRF-Token, Authorization")
		if ctx.Request.Method == "OPTIONS" {
			ctx.String(200, "ok")
			return
		}
		ctx.Next()
	})

	router.GET("/sayhi/:from", func(ctx *gin.Context) {
		from := ctx.Param("from")
		ctx.JSON(http.StatusOK, SayHelloResponse{Err:0, Content:fmt.Sprintf("get sayhi from %s", from)})
	})
	swagger.Swagger2(router,"/sayhi/{from}","get", &swagger.StructParam{
		JsonData:nil,
		ResponseData:&SayHelloResponse{},
		Description:"打招呼",
		Tags:[]string{"打招呼"},
		Summary:"打招呼",
	})

	router.POST("/sayhello", func(ctx *gin.Context) {
		req := SayHelloParam{}
		ctx.ShouldBindJSON(&req)

		ctx.JSON(http.StatusOK, SayHelloResponse{Err:0, Content:fmt.Sprintf("get sayhello from %s with %s", req.From, req.Content)})
	})
	swagger.Swagger2(router,"/sayhello","post", &swagger.StructParam{
		JsonData:&SayHelloParam{},
		ResponseData:&SayHelloResponse{},
		Description:"发送消息",
		Tags:[]string{"发送消息"},
		Summary:"发送消息",
	})
}