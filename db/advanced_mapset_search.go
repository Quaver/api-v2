package db

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"github.com/Quaver/api2/enums"
)

var advancedSearchFilterPattern = regexp.MustCompile(`^([A-Za-z]+)(<=|>=|=|<|>)(\S+)$`)
var advancedSearchFilterLikePattern = regexp.MustCompile(`^[A-Za-z]+(?:<=|>=|=|<|>).*$`)

type advancedRangeBound struct {
	floatValue float64
	intValue   int64
	inclusive  bool
}

type advancedRangeConstraint struct {
	integer bool
	lower   *advancedRangeBound
	upper   *advancedRangeBound
}

type parsedAdvancedSearch struct {
	text        string
	modes       []enums.GameMode
	statuses    []enums.RankedStatus
	ranges      map[string]advancedRangeConstraint
	tags        []string
	hasModes    bool
	hasStatuses bool
	clanRanked  bool
}

// ApplyAdvancedSearch parses advanced_search after normal query binding. It deliberately
// only runs for the public mapset-search endpoint, which calls this method.
func (options *ElasticMapsetSearchOptions) ApplyAdvancedSearch() error {
	expression := strings.TrimSpace(options.AdvancedSearch)
	if expression == "" {
		return nil
	}

	parsed, err := parseAdvancedSearch(expression)
	if err != nil {
		return err
	}

	options.Search = parsed.text
	options.advancedRanges = parsed.ranges
	options.advancedTags = parsed.tags

	if parsed.hasModes {
		options.Mode = parsed.modes
	}

	if parsed.hasStatuses {
		options.RankedStatus = parsed.statuses
	}

	if parsed.clanRanked {
		options.IsClanRanked = true
	}

	return nil
}

func parseAdvancedSearch(expression string) (*parsedAdvancedSearch, error) {
	parsed := &parsedAdvancedSearch{ranges: make(map[string]advancedRangeConstraint)}
	text := make([]string, 0)
	modeSeen := make(map[enums.GameMode]bool)
	statusSeen := make(map[enums.RankedStatus]bool)
	tagSeen := make(map[string]bool)

	for _, token := range strings.Fields(expression) {
		matches := advancedSearchFilterPattern.FindStringSubmatch(token)
		if matches == nil {
			if advancedSearchFilterLikePattern.MatchString(token) {
				return nil, fmt.Errorf("invalid advanced search filter %q", token)
			}

			text = append(text, token)
			continue
		}

		field, operator, value := strings.ToLower(matches[1]), matches[2], matches[3]
		switch field {
		case "t":
			if operator != "=" {
				return nil, fmt.Errorf("advanced search filter %q only supports =", field)
			}

			tag, err := advancedSearchTag(value)
			if err != nil {
				return nil, err
			}

			if !tagSeen[tag] {
				parsed.tags = append(parsed.tags, tag)
				tagSeen[tag] = true
			}
		case "k":
			if operator != "=" {
				return nil, fmt.Errorf("advanced search filter %q only supports =", field)
			}

			mode, err := gameModeFromKeyCount(value)
			if err != nil {
				return nil, err
			}

			parsed.hasModes = true
			if !modeSeen[mode] {
				parsed.modes = append(parsed.modes, mode)
				modeSeen[mode] = true
			}
		case "s":
			if operator != "=" {
				return nil, fmt.Errorf("advanced search filter %q only supports =", field)
			}
			if strings.EqualFold(value, "c") {
				parsed.clanRanked = true
				continue
			}

			status, err := rankedStatusFromAdvancedSearch(value)
			if err != nil {
				return nil, err
			}

			parsed.hasStatuses = true
			if !statusSeen[status] {
				parsed.statuses = append(parsed.statuses, status)
				statusSeen[status] = true
			}
		case "b", "d", "l", "ln", "pc", "lu":
			fieldName, integer := advancedRangeField(field)
			bound, err := parseAdvancedRangeBound(value, operator, integer)
			if err != nil {
				return nil, fmt.Errorf("invalid value for advanced search filter %q: %w", field, err)
			}

			constraint := parsed.ranges[fieldName]
			constraint.integer = integer
			if err := constraint.add(operator, bound); err != nil {
				return nil, fmt.Errorf("invalid advanced search filter %q: %w", field, err)
			}
			parsed.ranges[fieldName] = constraint
		default:
			return nil, fmt.Errorf("unknown advanced search filter %q", field)
		}
	}

	parsed.text = strings.Join(text, " ")
	return parsed, nil
}

func advancedSearchTag(value string) (string, error) {
	tag := strings.ToLower(value)
	tag = strings.NewReplacer("_", " ", "-", " ").Replace(tag)
	for _, validTag := range tagSearchTerms {
		if tag == validTag {
			return validTag, nil
		}
	}

	return "", fmt.Errorf("invalid tag %q", value)
}

func advancedRangeField(field string) (string, bool) {
	switch field {
	case "b":
		return "bpm", false
	case "d":
		return "difficulty_rating", false
	case "l":
		return "length", false
	case "ln":
		return "long_note_percentage", false
	case "pc":
		return "play_count", true
	case "lu":
		return "date_last_updated", true
	default:
		panic("unsupported advanced range field")
	}
}

func parseAdvancedRangeBound(value, operator string, integer bool) (advancedRangeBound, error) {
	bound := advancedRangeBound{inclusive: operator == "=" || operator == ">=" || operator == "<="}

	if integer {
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return advancedRangeBound{}, fmt.Errorf("must be an integer")
		}
		bound.intValue = parsed
		return bound, nil
	}

	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		return advancedRangeBound{}, fmt.Errorf("must be a finite number")
	}
	bound.floatValue = parsed
	return bound, nil
}

func (constraint *advancedRangeConstraint) add(operator string, bound advancedRangeBound) error {
	if operator == "=" {
		constraint.addLower(bound)
		constraint.addUpper(bound)
	} else if operator == ">" || operator == ">=" {
		constraint.addLower(bound)
	} else {
		constraint.addUpper(bound)
	}

	if constraint.lower == nil || constraint.upper == nil {
		return nil
	}

	comparison := constraint.compare(*constraint.lower, *constraint.upper)
	if comparison > 0 || comparison == 0 && (!constraint.lower.inclusive || !constraint.upper.inclusive) {
		return fmt.Errorf("range bounds do not overlap")
	}

	return nil
}

func (constraint *advancedRangeConstraint) addLower(bound advancedRangeBound) {
	if constraint.lower == nil || constraint.compare(bound, *constraint.lower) > 0 ||
		constraint.compare(bound, *constraint.lower) == 0 && !bound.inclusive && constraint.lower.inclusive {
		constraint.lower = &bound
	}
}

func (constraint *advancedRangeConstraint) addUpper(bound advancedRangeBound) {
	if constraint.upper == nil || constraint.compare(bound, *constraint.upper) < 0 ||
		constraint.compare(bound, *constraint.upper) == 0 && !bound.inclusive && constraint.upper.inclusive {
		constraint.upper = &bound
	}
}

func (constraint advancedRangeConstraint) compare(left, right advancedRangeBound) int {
	if constraint.integer {
		switch {
		case left.intValue < right.intValue:
			return -1
		case left.intValue > right.intValue:
			return 1
		default:
			return 0
		}
	}

	switch {
	case left.floatValue < right.floatValue:
		return -1
	case left.floatValue > right.floatValue:
		return 1
	default:
		return 0
	}
}

func gameModeFromKeyCount(value string) (enums.GameMode, error) {
	keyCount, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid game mode %q", value)
	}

	modes := map[int]enums.GameMode{
		1: enums.GameModeKeys1, 2: enums.GameModeKeys2, 3: enums.GameModeKeys3,
		4: enums.GameModeKeys4, 5: enums.GameModeKeys5, 6: enums.GameModeKeys6,
		7: enums.GameModeKeys7, 8: enums.GameModeKeys8, 9: enums.GameModeKeys9,
		10: enums.GameModeKeys10,
	}

	mode, exists := modes[keyCount]
	if !exists {
		return 0, fmt.Errorf("invalid game mode %q", value)
	}

	return mode, nil
}

func rankedStatusFromAdvancedSearch(value string) (enums.RankedStatus, error) {
	switch strings.ToLower(value) {
	case "r":
		return enums.RankedStatusRanked, nil
	case "u":
		return enums.RankedStatusUnranked, nil
	default:
		return 0, fmt.Errorf("invalid ranked status %q (expected r, u, or c)", value)
	}
}

func addSearchRangeQuery[T Number](boolQuery *BoolQuery, options *ElasticMapsetSearchOptions, field string, min T, max T) {
	if constraint, exists := options.advancedRanges[field]; exists {
		addAdvancedRangeQuery(boolQuery, field, constraint)
		return
	}

	addRangeQuery(boolQuery, field, min, max)
}

func addAdvancedRangeQuery(boolQuery *BoolQuery, field string, constraint advancedRangeConstraint) {
	rangeQuery := Range{}

	if constraint.lower != nil {
		value := advancedRangeValue(*constraint.lower, constraint.integer)
		if constraint.lower.inclusive {
			rangeQuery.Gte = value
		} else {
			rangeQuery.Gt = value
		}
	}

	if constraint.upper != nil {
		value := advancedRangeValue(*constraint.upper, constraint.integer)
		if constraint.upper.inclusive {
			rangeQuery.Lte = value
		} else {
			rangeQuery.Lt = value
		}
	}

	boolQuery.BoolQuery.Must = append(boolQuery.BoolQuery.Must, RangeCustom{
		Range: map[string]Range{field: rangeQuery},
	})
}

func advancedRangeValue(bound advancedRangeBound, integer bool) interface{} {
	if integer {
		return bound.intValue
	}

	return bound.floatValue
}
