package v1

import (
	"fmt"
	"github.com/Quaver/api2/config"
	"github.com/go-resty/resty/v2"
)

// UpdateElasticSearchMapset Makes an API call to v1 to update elastic search for a given mapset
func UpdateElasticSearchMapset(id int) error {
	resp, err := resty.New().R().
		SetHeader("Accept", "application/json").
		SetBody(map[string]interface{}{
			"key": config.Instance.ApiV1.SecretKey,
			"id":  id,
		}).Post(fmt.Sprintf("%v/v1/mapsets/elastic", config.Instance.ApiV1.Url))

	if err != nil {
		return err
	}

	if resp.IsError() {
		return fmt.Errorf("request to update elastic mapset (v1) failed: %v", resp.Error())
	}

	return nil
}
