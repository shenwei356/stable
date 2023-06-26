// Copyright © 2023 Wei Shen <shenwei356@gmail.com>
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.
package table

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/dustin/go-humanize"
)

// from https://github.com/tatsushid/go-prettytable
func convertToString(v interface{}, addComma bool) (string, error) {
	if addComma {
		switch vv := v.(type) {
		case fmt.Stringer:
			return vv.String(), nil
		case int:
			return humanize.Comma(int64(vv)), nil
		case int8:
			return humanize.Comma(int64(vv)), nil
		case int16:
			return humanize.Comma(int64(vv)), nil
		case int32:
			return humanize.Comma(int64(vv)), nil
		case int64:
			return humanize.Comma(vv), nil
		case uint:
			return humanize.Comma(int64(vv)), nil
		case uint8:
			return humanize.Comma(int64(vv)), nil
		case uint16:
			return humanize.Comma(int64(vv)), nil
		case uint32:
			return humanize.Comma(int64(vv)), nil
		case uint64:
			return humanize.Comma(int64(vv)), nil
		case float32:
			return humanize.Commaf(float64(vv)), nil
		case float64:
			return humanize.Commaf(float64(vv)), nil
		case bool:
			return strconv.FormatBool(vv), nil
		case string:
			return vv, nil
		case []byte:
			return string(vv), nil
		case []rune:
			return string(vv), nil
		default:
			return "", errors.New("can't convert the value")
		}
	}

	switch vv := v.(type) {
	case fmt.Stringer:
		return vv.String(), nil
	case int:
		return strconv.FormatInt(int64(vv), 10), nil
	case int8:
		return strconv.FormatInt(int64(vv), 10), nil
	case int16:
		return strconv.FormatInt(int64(vv), 10), nil
	case int32:
		return strconv.FormatInt(int64(vv), 10), nil
	case int64:
		return strconv.FormatInt(vv, 10), nil
	case uint:
		return strconv.FormatUint(uint64(vv), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(vv), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(vv), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(vv), 10), nil
	case uint64:
		return strconv.FormatUint(vv, 10), nil
	case float32:
		return strconv.FormatFloat(float64(vv), 'g', -1, 32), nil
	case float64:
		return strconv.FormatFloat(vv, 'g', -1, 64), nil
	case bool:
		return strconv.FormatBool(vv), nil
	case string:
		return vv, nil
	case []byte:
		return string(vv), nil
	case []rune:
		return string(vv), nil
	default:
		return "", errors.New("can't convert the value")
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
