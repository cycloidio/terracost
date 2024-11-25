package testutil

import (
	"encoding/json"
	"fmt"
)

func JSON(in any) string {
	ret, err := json.MarshalIndent(in, "", "  ")
	if err != nil {
		return fmt.Sprintf("%+v", err)
	}
	return string(ret)
}
