package entities

import (
	"encoding/json"
	"fmt"
)

type OptionType string

const (
	OptionTypeReference        OptionType = "reference"
	OptionTypeChoice           OptionType = "choice"
	OptionTypeMultiple         OptionType = "multiple"
	OptionTypeCountedReference OptionType = "counted_reference"
)

type ChoiceStatus string

const (
	ChoiceStatusUnset    ChoiceStatus = ""
	ChoiceStatusActive   ChoiceStatus = "active"
	ChoiceStatusInactive ChoiceStatus = "inactive"
	ChoiceStatusSelected ChoiceStatus = "selected"
)

type SelectOuput struct {
	Option  Option
	HasMore bool
}

type ChoiceType string

const (
	ChoiceTypeUnset       ChoiceType = ""
	ChoiceTypeProficiency ChoiceType = "proficiency"
	ChoiceTypeLanguage    ChoiceType = "language"
	ChoiceTypeEquipment   ChoiceType = "equipment"
)

type Choice struct {
	Name     string       `json:"name"`
	Type     ChoiceType   `json:"type"`
	Key      string       `json:"key"`
	Status   ChoiceStatus `json:"status"`
	Selected bool         `json:"selected"`
	Count    int          `json:"count"`
	Options  []Option     `json:"options"`
}

func (o *Choice) GetOptionType() OptionType {
	return OptionTypeChoice
}

func (o *Choice) GetName() string {
	return o.Name
}

func (o *Choice) GetKey() string {
	return o.Key
}

func (o *Choice) GetStatus() ChoiceStatus {
	return o.Status
}

func (o *Choice) SetStatus(status ChoiceStatus) {
	o.Status = status
}

// Select selects an option by key
// sets the current choice to Active if there are remaining opiotns to be chosen, otherwise marks the top level choice as Selected
func (o *Choice) Select(key string) *SelectOuput {
	selected := &SelectOuput{}
	selectedCount := 0

	for _, option := range o.Options {
		if option.GetStatus() == ChoiceStatusSelected {
			selectedCount++
			continue // if this option is selected we don't need to do anything
		}

		if option.GetKey() == key {
			switch option.GetOptionType() {
			case OptionTypeReference, OptionTypeCountedReference:
				option.SetStatus(ChoiceStatusSelected)
				selected.Option = option
			case OptionTypeChoice, OptionTypeMultiple:
				option.SetStatus(ChoiceStatusActive)
				selected.Option = option
				selected.HasMore = true
			}

			selectedCount++

			break
		}

		if option.GetOptionType() == OptionTypeChoice {
			choiceOption := option.(*Choice)
			choice := choiceOption.Select(key)
			if choice.Option != nil {
				selected = choice
				if selected.Option.GetStatus() == ChoiceStatusSelected {
					selectedCount++
				}
			}
			// TODO: if the selected status is selected what should we do?
		}

		if option.GetOptionType() == OptionTypeMultiple {
			multipleOption := option.(*MultipleOption)
			multiple := multipleOption.Select(key)
			if multiple.Option != nil {
				selected = multiple
				if selected.Option.GetStatus() == ChoiceStatusSelected {
					selectedCount++
				}
			}
		}

		if option.GetKey() == key {
			selected.Option = option

			selectedCount++
		}
	}

	if selectedCount == o.Count && !selected.HasMore {
		o.Status = ChoiceStatusSelected
	} else if selectedCount > 0 {
		o.Status = ChoiceStatusActive
	}

	return selected
}

type Option interface {
	GetOptionType() OptionType
	GetName() string
	GetKey() string
	GetStatus() ChoiceStatus
	SetStatus(ChoiceStatus)
	Select(key string) *SelectOuput
}

type CountedReferenceOption struct {
	Status    ChoiceStatus   `json:"status"`
	Count     int            `json:"count"`
	Reference *ReferenceItem `json:"reference"`
}

func (o *CountedReferenceOption) Select(key string) *SelectOuput {
	o.Status = ChoiceStatusSelected

	return &SelectOuput{Option: o, HasMore: false}
}

func (o *CountedReferenceOption) GetOptionType() OptionType {
	return OptionTypeCountedReference
}

func (o *CountedReferenceOption) GetName() string {
	return o.Reference.Name
}

func (o *CountedReferenceOption) GetStatus() ChoiceStatus {
	return o.Status
}

func (o *CountedReferenceOption) SetStatus(status ChoiceStatus) {
	o.Status = status
}

func (o *CountedReferenceOption) GetKey() string {
	return o.Reference.Key
}

type ReferenceOption struct {
	Status    ChoiceStatus   `json:"status"`
	Reference *ReferenceItem `json:"reference"`
}

func (o *ReferenceOption) Select(key string) *SelectOuput {
	o.Status = ChoiceStatusSelected

	return &SelectOuput{Option: o, HasMore: false}
}

func (o *ReferenceOption) GetOptionType() OptionType {
	return OptionTypeReference
}

func (o *ReferenceOption) GetName() string {
	return o.Reference.Name
}

func (o *ReferenceOption) GetKey() string {
	return o.Reference.Key
}

func (o *ReferenceOption) GetStatus() ChoiceStatus {
	return o.Status
}

func (o *ReferenceOption) SetStatus(status ChoiceStatus) {
	o.Status = status
}

type MultipleOption struct {
	Status ChoiceStatus `json:"status"`
	Key    string       `json:"key"`
	Name   string       `json:"name"`
	Items  []Option     `json:"items"`
}

func (o *MultipleOption) Select(key string) *SelectOuput {
	totalCount := 0
	selected := &SelectOuput{}

	// Go through all the items and select the one that matches the key
	// keep track of how many have been selected
	for idx, item := range o.Items {
		if item.GetStatus() == ChoiceStatusSelected {
			totalCount++
			continue
		}

		if item.GetKey() == key {
			totalCount++

			if item.GetStatus() != ChoiceStatusSelected {
				selected.Option = item
				item.SetStatus(ChoiceStatusSelected)

				o.Status = ChoiceStatusActive

				if idx < len(o.Items)-1 {
					selected.HasMore = true
				}

				break
			}
		}

		current := item.Select(key)
		if current != nil {
			selected = current
			if selected.Option.GetStatus() == ChoiceStatusSelected {
				totalCount++
			}

			o.Status = ChoiceStatusActive

			break
		}
	}

	// If they have all been selected we will mark the top level option as selected
	if totalCount == len(o.Items) {
		o.Status = ChoiceStatusSelected
		selected.HasMore = false
	}

	return selected
}
func (o *MultipleOption) GetOptionType() OptionType {
	return OptionTypeMultiple
}

func (o *MultipleOption) GetName() string {
	return o.Name
}

func (o *MultipleOption) GetKey() string {
	return o.Key
}

func (o *MultipleOption) GetStatus() ChoiceStatus {
	return o.Status
}

func (o *MultipleOption) SetStatus(status ChoiceStatus) {
	o.Status = status
}

// UnmarshalJSON implements custom JSON unmarshaling for Choice
func (c *Choice) UnmarshalJSON(data []byte) error {
	// Create an auxiliary struct to avoid infinite recursion
	type Aux struct {
		Name     string          `json:"name"`
		Type     ChoiceType      `json:"type"`
		Key      string          `json:"key"`
		Status   ChoiceStatus    `json:"status"`
		Selected bool            `json:"selected"`
		Count    int             `json:"count"`
		Options  json.RawMessage `json:"options"`
	}

	var aux Aux
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	c.Name = aux.Name
	c.Type = aux.Type
	c.Key = aux.Key
	c.Status = aux.Status
	c.Selected = aux.Selected
	c.Count = aux.Count

	// Handle options which could be an array or object
	if len(aux.Options) > 0 {
		// First try to unmarshal as an array of options
		var optionsArray []json.RawMessage
		if err := json.Unmarshal(aux.Options, &optionsArray); err == nil {
			// It's an array, process each option
			c.Options = make([]Option, 0, len(optionsArray))
			for _, optData := range optionsArray {
				opt, err := unmarshalOption(optData)
				if err != nil {
					return fmt.Errorf("failed to unmarshal option: %w", err)
				}
				if opt != nil {
					c.Options = append(c.Options, opt)
				}
			}
		} else {
			// If it's not an array, it might be an object (which we'll treat as a single option)
			// For now, we'll just ignore it as it seems the API might return different formats
			c.Options = []Option{}
		}
	}

	return nil
}

// unmarshalOption unmarshals a single option based on its type
func unmarshalOption(data []byte) (Option, error) {
	// First, determine the type of option
	var typeCheck struct {
		Type      string          `json:"type"`
		Reference json.RawMessage `json:"reference"`
		Count     *int            `json:"count"`
		Items     json.RawMessage `json:"items"`
	}

	if err := json.Unmarshal(data, &typeCheck); err != nil {
		return nil, err
	}

	// Check if it has a reference field
	if len(typeCheck.Reference) > 0 {
		if typeCheck.Count != nil {
			// It's a CountedReferenceOption
			var opt CountedReferenceOption
			if err := json.Unmarshal(data, &opt); err != nil {
				return nil, err
			}
			return &opt, nil
		} else {
			// It's a ReferenceOption
			var opt ReferenceOption
			if err := json.Unmarshal(data, &opt); err != nil {
				return nil, err
			}
			return &opt, nil
		}
	}

	// Check if it has items field (MultipleOption)
	if len(typeCheck.Items) > 0 {
		var opt MultipleOption
		if err := json.Unmarshal(data, &opt); err != nil {
			return nil, err
		}
		return &opt, nil
	}

	// Check if it's a Choice based on type field
	if typeCheck.Type != "" {
		var opt Choice
		if err := json.Unmarshal(data, &opt); err != nil {
			return nil, err
		}
		return &opt, nil
	}

	// If we can't determine the type, return nil
	return nil, nil
}

// UnmarshalJSON implements custom JSON unmarshaling for MultipleOption
func (m *MultipleOption) UnmarshalJSON(data []byte) error {
	// Create an auxiliary struct to avoid infinite recursion
	type Aux struct {
		Status ChoiceStatus    `json:"status"`
		Key    string          `json:"key"`
		Name   string          `json:"name"`
		Items  json.RawMessage `json:"items"`
	}

	var aux Aux
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	m.Status = aux.Status
	m.Key = aux.Key
	m.Name = aux.Name

	// Handle items which should be an array of options
	if len(aux.Items) > 0 {
		var itemsArray []json.RawMessage
		if err := json.Unmarshal(aux.Items, &itemsArray); err != nil {
			return fmt.Errorf("failed to unmarshal items array: %w", err)
		}

		m.Items = make([]Option, 0, len(itemsArray))
		for _, itemData := range itemsArray {
			item, err := unmarshalOption(itemData)
			if err != nil {
				return fmt.Errorf("failed to unmarshal item: %w", err)
			}
			if item != nil {
				m.Items = append(m.Items, item)
			}
		}
	}

	return nil
}
