package httpx

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

func EvaluateExpression(query string) (string, error) {
	endpoint := fmt.Sprintf("https://evaluate-expression.p.rapidapi.com/?expression=%s", url.QueryEscape(query))
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("x-rapidapi-host", "evaluate-expression.p.rapidapi.com")
	req.Header.Set("x-rapidapi-key", "cf9e67ea99mshecc7e1ddb8e93d1p1b9e04jsn3f1bb9103c3f")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	result := string(body)
	if result == "" {
		return "", fmt.Errorf("invalid Math Expression")
	}

	return result, nil
}
