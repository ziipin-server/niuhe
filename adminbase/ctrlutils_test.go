package adminbase

import "testing"

func TestPascalToSnake(t *testing.T) {
	if orig := "AppId"; pascalToSnake(orig) != "app_id" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "Name"; pascalToSnake(orig) != "name" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "Name1"; pascalToSnake(orig) != "name_1" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
	if orig := "UserName1984"; pascalToSnake(orig) != "user_name_1984" {
		t.Error("pascalToSnake error", orig, pascalToSnake(orig))
	}
}
