package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func HomeHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/json")
}

func XmlHandler(c *gin.Context) {
	c.XML(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/xml")
}

func TextHandler(c *gin.Context) {
	c.String(http.StatusOK, "Hello World")

	c.Request.Response.Header.Add("Content-Type", "text/plain")
}

func YamlHandler(c *gin.Context) {
	c.YAML(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/x-yaml")
}

func ProtobufHandler(c *gin.Context) {
	c.ProtoBuf(http.StatusOK, gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "application/x-protobuf")
}

func ServerSideEventsHandler(c *gin.Context) {
	c.SSEvent("message", gin.H{
		"message": "Hello World",
	})

	c.Request.Response.Header.Add("Content-Type", "text/event-stream")
}

func SecretHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "The secret ingredient is ...",
	})

	c.Request.Response.Header.Add("Content-Type", "application/json")
}
