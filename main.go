package swagger

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/url"
	"fmt"
	"strings"
	"encoding/json"
)

type Config struct {
	// api url，内部自动判断http，https
	Url string

	// "api前缀，例如/api/v1"，默认为空
	BasePath string

	// swagger文档标题
	Title string

	// swagger文档描述
	Description string

	// 文档版本
	DocVersion string

	// swagger ui的地址
	SwaggerUiUrl string

	// 文档Url地址，例如开发服务器http://baidu.com
	// 如果本值是doc，那么http://baidu.com/doc就可以重定向到@SwaggerUiUrl
	SwaggerUrlPrefix string

	// swagger文档的地址，用于调试，release直接打包到二进制里面。默认为空
	DocFilePath string

	// 用于支持swagger ui认证头的参数
	Headers []SecurityDefinition

	// 是否调试模式
	Debug bool
}

func (this *Config) initDefault() {
	if len(this.Title) == 0 {
		this.Title = "Swagger document"
	}
	if len(this.Description) == 0 {
		this.Description = "Swagger document description"
	}
	if len(this.DocVersion) == 0 {
		this.DocVersion = "0.0.1"
	}
	if len(this.SwaggerUiUrl) == 0 {
		// http://swagger.qiujinwu.com
		this.SwaggerUiUrl = "http://petstore.swagger.io/"
	}
	if len(this.SwaggerUrlPrefix) == 0 {
		this.SwaggerUrlPrefix = "apidoc"
	}
}

func InitializeApiRoutes(grouter *gin.Engine, config *Config, docLoader DocLoader) {
	if gDefaultOption != nil {
		panic("swagger inited yet")
		return
	}

	if config == nil || docLoader == nil {
		panic("invalid swagger parameter")
	}
	config.initDefault()
	gDefaultOption = newOptions(config)
	gDefaultOption.docLoader = docLoader

	grouter.GET("/"+config.SwaggerUrlPrefix+"/spec/:group", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		lang := c.Query("lang")
		fmt.Println("lang", "--", lang)
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
			fmt.Println("Accept-Language", "--", lang)
		}

		apiGroupName := c.Param("group")
		apiGroupName = strings.TrimLeft(apiGroupName, "/")
		apiGroupName = strings.TrimRight(apiGroupName, "/")

		swaggerData1 := gDefaultOption.swaggerData
		if v, ok := gDefaultOption.swaggerDataMap[apiGroupName]; ok {
			swaggerData1 = v
		}

		headersDef := make(map[string]SecurityDefinition)
		if len(config.Headers) > 0 {
			for _, v := range config.Headers {
				key := v.Type
				v.In = "header"
				v.Type = "apiKey"
				headersDef[key] = v
			}
		}

		url := TranslateText(config.Url, lang)
		var (
			scheme string
			host string
		)
		const (
			httpPrefix = "http://"
			httpsPrefix = "https://"
		)
		if strings.HasPrefix(url, httpPrefix) {
			scheme = "http"
			host = url[len(httpPrefix):]
		} else if strings.HasPrefix(url, httpsPrefix) {
			scheme = "https"
			host = url[len(httpsPrefix):]
		} else {
			scheme = "http"
			host = url
		}

		response := gin.H{
			"schemes":[]string{scheme},
			"host": host,
			"basePath": config.BasePath,
			"swagger":  swaggerVersion,
			"info": struct {
				Description string `json:"description"`
				Title       string `json:"title"`
				Version     string `json:"version"`
			}{
				Description: config.Description,
				Title:       config.Title,
				Version:     config.DocVersion,
			},
			//"definition":          struct{}{},
			"paths":               swaggerData1,
			"securityDefinitions": headersDef,
		}

		//c.JSON(http.StatusOK, response)

		data, _ := json.Marshal(response)
		text := TranslateText(string(data), lang)
		c.String(http.StatusOK, "%s", text)

	})

	grouter.GET("/"+config.SwaggerUrlPrefix, func(c *gin.Context) {
		scheme := "http://"
		if c.Request.TLS != nil {
			scheme = "https://"
		}
		host := scheme + c.Request.Host + "/" + config.SwaggerUrlPrefix + "/spec"
		host = config.SwaggerUiUrl + "?url=" + url.QueryEscape(host)
		c.Redirect(http.StatusFound, host)
	})
}

func AddGroupOption(groupName string, config *Config, docLoader DocLoader) {
	//if gOption != nil {
	//	panic("swagger inited yet")
	//	return
	//}

	if config == nil || docLoader == nil {
		panic("invalid swagger parameter")
	}
	config.initDefault()
	option := newOptions(config)
	option.docLoader = docLoader

	gGroupOptions[groupName] = option
}

func InitializeApiRoutesByGroup(grouter *gin.Engine, urlPrefix string) {
	grouter.GET("/" + urlPrefix + "/spec/*group", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		lang := c.Query("lang")
		fmt.Println("lang", "--", lang)
		if lang == "" {
			lang = c.GetHeader("Accept-Language")
			fmt.Println("Accept-Language", "--", lang)
		}

		apiGroupName := c.Param("group")
		apiGroupName = strings.TrimLeft(apiGroupName, "/")
		apiGroupName = strings.TrimRight(apiGroupName, "/")

		var (
			option *options
			exist bool
		)
		option, exist = gGroupOptions[apiGroupName]
		if !exist {
			c.JSON(http.StatusOK, struct {
				ErrMsg string `json:"errmsg"`
			}{ErrMsg: fmt.Sprintf("Not find group option %s", apiGroupName)})
			return
		}

		swaggerData1 := option.swaggerData
		if v, ok := option.swaggerDataMap[apiGroupName]; ok {
			swaggerData1 = v
		}

		headersDef := make(map[string]SecurityDefinition)
		if len(option.config.Headers) > 0 {
			for _, v := range option.config.Headers {
				key := v.Type
				v.In = "query"
				v.Type = "basic"
				headersDef[key] = v
			}
		}

		url := TranslateText(option.config.Url, lang)
		var (
			scheme string
			host string
		)
		const (
			httpPrefix = "http://"
			httpsPrefix = "https://"
		)
		if strings.HasPrefix(url, httpPrefix) {
			scheme = "http"
			host = url[len(httpPrefix):]
		} else if strings.HasPrefix(url, httpsPrefix) {
			scheme = "https"
			host = url[len(httpsPrefix):]
		} else {
			scheme = "http"
			host = url
		}

		response := gin.H{
			"schemes":[]string{scheme},
			"host": host,
			"basePath": option.config.BasePath,
			"swagger":  swaggerVersion,
			"info": struct {
				Description string `json:"description"`
				Title       string `json:"title"`
				Version     string `json:"version"`
			}{
				Description: option.config.Description,
				Title:       option.config.Title,
				Version:     option.config.DocVersion,
			},
			//"definition":          struct{}{},
			"paths":               swaggerData1,
			"securityDefinitions": headersDef,
		}

		//c.JSON(http.StatusOK, response)

		data, _ := json.Marshal(response)
		text := TranslateText(string(data), lang)
		c.String(http.StatusOK, "%s", text)

	})

	grouter.GET("/"+urlPrefix, func(c *gin.Context) {
		scheme := "http://"
		if c.Request.TLS != nil {
			scheme = "https://"
		}
		host := scheme + c.Request.Host + "/" + urlPrefix + "/spec"
		host = urlPrefix + "?url=" + url.QueryEscape(host)
		c.Redirect(http.StatusFound, host)
	})
}

