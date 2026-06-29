package niuhe

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestContextYAML(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	ginCtx, _ := gin.CreateTestContext(w)
	ctx := newContext(ginCtx, nil)

	ctx.YAML(201, gin.H{"message": "ok"})

	if w.Code != 201 {
		t.Fatalf("expected status 201, got %d", w.Code)
	}
	if w.Body.Len() == 0 {
		t.Fatal("expected YAML response body")
	}
}
