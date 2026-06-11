package db

import (
	"encoding/json"
	"fmt"
)

func (s *ChatSide) UnmarshalJSON(data []byte) error {

	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	switch ChatSide(v) {
	case ChatSideBottom, ChatSideTop:
		*s = ChatSide(v)
		return nil
	default:
		return fmt.Errorf("invalid chat side: %s", v)
	}
}

func (s *Size) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	switch Size(v) {
	case SizeSm, SizeMd, SizeLg:
		*s = Size(v)
		return nil
	default:
		return fmt.Errorf("invalid size: %s", v)
	}
}

func (t *Theme) UnmarshalJSON(data []byte) error {
	var v string
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	switch Theme(v) {
	case ThemeSystem, ThemeLight, ThemeDark:
		*t = Theme(v)
		return nil
	default:
		return fmt.Errorf("invalid theme: %s", v)
	}
}
