package device

import (
	"errors"
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"strings"

	. "github.com/mchubar/devicedetector/parser"
)

const UnknownBrand = "Unknown"

type DeviceMatchResult struct {
	Type  string `yaml:"type"`
	Model string `yaml:"model"`
	Brand string `yaml:"brand"`
}

type DeviceParser interface {
	PreMatch(string) bool
	Parse(string) *DeviceMatchResult
}

type Model struct {
	Regular `yaml:",inline" json:",inline"`
	Model   string `yaml:"model" json:"model"`
	Device  string `yaml:"device" json:"device"` //mobile
	Brand   string `yaml:"brand" json:"brand"`   //mobile
}

type DeviceReg struct {
	Brand   string
	Regular `yaml:",inline" json:",inline"`
	Model   string   `yaml:"model" json:"model"`
	Device  string   `yaml:"device" json:"device"`
	Models  []*Model `yaml:"models" json:"models"`
}

type DeviceParserAbstract struct {
	Regexes      []*DeviceReg
	overAllMatch Regular
}

func (d *DeviceParserAbstract) Load(file string) error {
	var v map[string]*DeviceReg
	err := ReadYamlFile(file, &v)

	m := yaml.MapSlice{}

	data, err := ioutil.ReadFile(file)
	if err != nil {
		return errors.New("not exists:" + file)
	}

	err = yaml.Unmarshal([]byte(data), &m)

	if err != nil {
		return err
	}

	for _, item := range v {
		item.Compile()
		for _, m := range item.Models {
			m.Compile()
		}
	}

	for _,item := range m{
		brand := fmt.Sprintf("%v", item.Key)
		regex := v[brand]
		regex.Brand = brand
		d.Regexes = append(d.Regexes, regex)
	}

	return nil
}

func (d *DeviceParserAbstract) PreMatch(ua string) bool {
	if d.overAllMatch.Regexp == nil {
		count := len(d.Regexes)
		if count == 0 {
			return false
		}
		sb := strings.Builder{}
		sb.WriteString(d.Regexes[count-1].Regex)
		for i := count - 2; i >= 0; i-- {
			sb.WriteString("|")
			sb.WriteString(d.Regexes[i].Regex)
		}
		d.overAllMatch.Regex = sb.String()
		d.overAllMatch.Compile()
	}
	r := d.overAllMatch.IsMatchUserAgent(ua)
	return r
}

func (d *DeviceParserAbstract) Parse(ua string) *DeviceMatchResult {
	var regex *DeviceReg
	var brand string
	var matches []string

	count := len(d.Regexes)
	for i := 0 ; i < count; i++ {
		regex = d.Regexes[i]
		brand = regex.Brand
		matches = regex.MatchUserAgent(ua)
		if len(matches) > 0 {
			break
		}
	}

	if regex == nil || len(matches) == 0 {
		return nil
	}

	r := &DeviceMatchResult{
		Type: regex.Device,
	}
	if brand != UnknownBrand {
		brandId := FindBrand(brand)
		if brandId == "" {
			return nil
		}
		r.Brand = brandId
	}

	if regex.Model != "" {
		r.Model = BuildModel(regex.Model, matches)
	}

	for _, modelRegex := range regex.Models {
		modelMatches := modelRegex.MatchUserAgent(ua)
		if len(modelMatches) > 0 {
			r.Model = strings.TrimSpace(BuildModel(modelRegex.Model, modelMatches))
			if modelRegex.Brand != "" {
				if brandId := FindBrand(modelRegex.Brand); brandId != "" {
					r.Brand = brandId
				}
			}
			if modelRegex.Device != "" {
				r.Type = modelRegex.Device
			}
			return r
		}
	}
	return r
}
