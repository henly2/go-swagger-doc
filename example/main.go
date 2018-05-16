package main

import(
	"github.com/gin-gonic/gin"
	"net/http"
	"github.com/qjw/go-swagger-doc"
	"fmt"
	"github.com/gin-contrib/cors"
)

type SayHelloParam struct {
	From   		string `json:"from"`
	Content  	string `json:"content"`
}

type SayHelloResponse struct {
	Err   	 int `json:"err"`
	Content  string `json:"content"`
}

func DocLoader(key string) ([]byte, error){
	fmt.Println("key:", key)
	return []byte("what"), nil
}

func main(){
	engine := gin.Default()
	engine.Use(cors.New(cors.Config{
		AllowAllOrigins:true,
		AllowMethods:     []string{"POST", "GET", "OPTIONS", "PUT", "DELETE"},
		AllowHeaders:     []string{"Authorization", "X-Requested-With", "X_Requested_With", "Content-Type", "Access-Token", "Accept-Language"},
		//AllowOrigins:     []string{"*"},
		//AllowCredentials: true,
		//AllowOriginFunc: func(origin string) bool {
		//	return true;//origin == "https://github.com"
		//},
		//MaxAge: 12 * time.Hour,
	}))

	config := swagger.Config{}
	swagger.InitializeApiRoutes(engine, &config, DocLoader)

	router := engine.Group("/api", func(ctx *gin.Context) {

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

	engine.Run(":8044")
}
