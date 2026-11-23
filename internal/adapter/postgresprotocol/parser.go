package postgresprotocol

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
)

type Message struct {
	Type byte
	Body []byte
}

type Parser struct{ buf []byte }

func (p *Parser) Feed(data []byte) ([]Message, error) {
	p.buf = append(p.buf, data...)
	var result []Message
	for {
		if len(p.buf) < 5 {
			break
		}
		length := int(binary.BigEndian.Uint32(p.buf[1:5]))
		if length < 4 || length > 16*1024*1024 {
			return nil, fmt.Errorf("invalid postgres message length %d", length)
		}
		if len(p.buf) < length+1 {
			break
		}
		body := append([]byte(nil), p.buf[5:length+1]...)
		result = append(result, Message{Type: p.buf[0], Body: body})
		p.buf = p.buf[length+1:]
	}
	return result, nil
}

func CString(data []byte, offset *int) (string, error) {
	if *offset >= len(data) {
		return "", errors.New("missing postgres cstring")
	}
	end := bytes.IndexByte(data[*offset:], 0)
	if end < 0 {
		return "", errors.New("unterminated postgres cstring")
	}
	value := string(data[*offset : *offset+end])
	*offset += end + 1
	return value, nil
}

type ParseMessage struct{ Name, Query string }

func ParseParse(m Message) (ParseMessage, error) {
	if m.Type != 'P' {
		return ParseMessage{}, errors.New("not a Parse message")
	}
	o := 0
	name, err := CString(m.Body, &o)
	if err != nil {
		return ParseMessage{}, err
	}
	query, err := CString(m.Body, &o)
	if err != nil {
		return ParseMessage{}, err
	}
	return ParseMessage{Name: name, Query: query}, nil
}

type BindMessage struct {
	Portal, Statement string
	Parameters        [][]byte
	Formats           []int16
}

func ParseBind(m Message) (BindMessage, error) {
	if m.Type != 'B' {
		return BindMessage{}, errors.New("not a Bind message")
	}
	o := 0
	portal, err := CString(m.Body, &o)
	if err != nil {
		return BindMessage{}, err
	}
	statement, err := CString(m.Body, &o)
	if err != nil {
		return BindMessage{}, err
	}
	formatCount, err := readInt16(m.Body, &o)
	if err != nil {
		return BindMessage{}, err
	}
	formats := make([]int16, formatCount)
	for i := range formats {
		formats[i], err = readInt16(m.Body, &o)
		if err != nil {
			return BindMessage{}, err
		}
	}
	count, err := readInt16(m.Body, &o)
	if err != nil {
		return BindMessage{}, err
	}
	params := make([][]byte, count)
	for i := range params {
		length, e := readInt32(m.Body, &o)
		if e != nil {
			return BindMessage{}, e
		}
		if length == -1 {
			continue
		}
		if length < 0 || o+int(length) > len(m.Body) {
			return BindMessage{}, errors.New("invalid Bind parameter")
		}
		params[i] = append([]byte(nil), m.Body[o:o+int(length)]...)
		o += int(length)
	}
	return BindMessage{Portal: portal, Statement: statement, Parameters: params, Formats: formats}, nil
}

func CommandTag(m Message) (string, error) {
	if m.Type != 'C' {
		return "", errors.New("not CommandComplete")
	}
	o := 0
	return CString(m.Body, &o)
}
func ErrorFields(m Message) (map[byte]string, error) {
	if m.Type != 'E' {
		return nil, errors.New("not ErrorResponse")
	}
	result := map[byte]string{}
	o := 0
	for o < len(m.Body) && m.Body[o] != 0 {
		code := m.Body[o]
		o++
		value, err := CString(m.Body, &o)
		if err != nil {
			return nil, err
		}
		result[code] = value
	}
	return result, nil
}

func readInt16(data []byte, offset *int) (int16, error) {
	if *offset+2 > len(data) {
		return 0, errors.New("short int16")
	}
	v := int16(binary.BigEndian.Uint16(data[*offset:]))
	*offset += 2
	return v, nil
}
func readInt32(data []byte, offset *int) (int32, error) {
	if *offset+4 > len(data) {
		return 0, errors.New("short int32")
	}
	v := int32(binary.BigEndian.Uint32(data[*offset:]))
	*offset += 4
	return v, nil
}
