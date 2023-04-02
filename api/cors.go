package api

import (
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
)

func HandleCorsRequestHandler(config ApiConfig, c *gin.Context) {
	HandleCorsRequest(config, c, nil)
}

func HandleCorsRequest(config ApiConfig, c *gin.Context, client *http.Client) {
	corsURL := c.Query("url")
	if corsURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "url query parameter is required"})
		return
	}

	// Create a new request to the target URL
	targetReq, err := http.NewRequest(c.Request.Method, corsURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Forward headers from the client request to the target request
	for name, values := range c.Request.Header {
		for _, value := range values {
			targetReq.Header.Add(name, value)
		}
	}

	// Make a request to the target URL
	if client == nil {
		client = &http.Client{}
	}
	targetResp, err := client.Do(targetReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer targetResp.Body.Close()

	// Copy the response headers from the target URL to the client response
	for name, values := range targetResp.Header {
		for _, value := range values {
			c.Header(name, value)
		}
	}

	// Copy the response body from the target URL to the client response
	c.Status(targetResp.StatusCode)
	io.Copy(c.Writer, targetResp.Body)
}
