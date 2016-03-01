package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

var (
	iniComment      = []byte{';'} // comment
	iniEmpty        = []byte{}    // empty
	iniEqual        = []byte{'='} // equal
	iniDQuote       = []byte{'"'} // quote
	iniSectionStart = []byte{'['} // section start
	iniSectionEnd   = []byte{']'} // section end
	iniLineBreak    = "\n"        // new line
)

type IniConfigAdapter struct {
}

func (ini *IniConfigAdapter) ParseFile(name string) (Configurer, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	cfg := &IniConfigurer{
		file.Name(),
		make(map[string]map[string]string),
		make(map[string]string),
		make(map[string]string),
		sync.RWMutex{},
	}
	cfg.Lock()
	defer cfg.Unlock()

	var commentBuf bytes.Buffer
	buf := bufio.NewReader(file)

	// check the BOM
	head, err := buf.Peek(3)
	if err == nil && head[0] == 239 && head[1] == 187 && head[2] == 191 {
		for i := 1; i <= 3; i++ {
			buf.ReadByte()
		}
	}

	var section string
	for {
		line, _, err := buf.ReadLine()

		if err == io.EOF {
			break
		}

		if bytes.Equal(line, iniEmpty) {
			continue
		}

		line = bytes.TrimSpace(line)

		if bytes.HasPrefix(line, iniComment) {
			line = bytes.TrimLeft(line, string(iniComment))
			line = bytes.TrimLeftFunc(line, unicode.IsSpace)
			commentBuf.Write(line)
			commentBuf.WriteByte('\n')
			continue
		}

		if bytes.HasPrefix(line, iniSectionStart) && bytes.HasSuffix(line, iniSectionEnd) {
			section = strings.ToLower(string(line[1 : len(line)-1])) // section name case insensitive
			if commentBuf.Len() > 0 {
				cfg.sectionComment[section] = commentBuf.String()
				commentBuf.Reset()
			}
			if _, ok := cfg.data[section]; !ok {
				cfg.data[section] = make(map[string]string)
			}
			continue
		}

		if _, ok := cfg.data[section]; !ok {
			cfg.data[section] = make(map[string]string)
		}
		keyValue := bytes.SplitN(line, iniEqual, 2)

		key := string(bytes.TrimSpace(keyValue[0])) // key name case insensitive
		key = strings.ToLower(key)

		if len(keyValue) != 2 {
			return nil, errors.New("read the content error: \"" + string(line) + "\", should key = val")
		}
		val := bytes.TrimSpace(keyValue[1])
		if bytes.HasPrefix(val, iniDQuote) {
			val = bytes.Trim(val, `"`)
		}

		cfg.data[section][key] = string(val)
		if commentBuf.Len() > 0 {
			cfg.keyComment[section+"."+key] = commentBuf.String()
			commentBuf.Reset()
		}
	}
	return cfg, nil
}

func (ini *IniConfigAdapter) ParseData(data []byte) (Configurer, error) {
	tmpName := path.Join(os.TempDir(), "golite", fmt.Sprintf("%d", time.Now().Nanosecond()))
	os.MkdirAll(path.Dir(tmpName), os.ModePerm)
	if err := ioutil.WriteFile(tmpName, data, 0655); err != nil {
		return nil, err
	}
	defer os.Remove(tmpName)
	return ini.ParseFile(tmpName)
}

type IniConfigurer struct {
	filename       string
	data           map[string]map[string]string // section.key = val
	sectionComment map[string]string            // section : comment
	keyComment     map[string]string
	sync.RWMutex
}

func (c *IniConfigurer) GetBool(key string, v ...bool) (bool, error) {
	var (
		defval bool
		bdef   bool
	)

	if len(v) > 0 {
		defval = v[0]
		bdef = true
	}

	if val, err := strconv.ParseBool(c.getdata(key)); err != nil {
		if bdef {
			return defval, nil
		} else {
			return defval, err
		}
	} else {
		return val, nil
	}
}

func (c *IniConfigurer) GetFloat(key string, v ...float64) (float64, error) {
	var (
		defval float64
		bdef   bool
	)

	if len(v) > 0 {
		defval = v[0]
		bdef = true
	}

	if val, err := strconv.ParseFloat(c.getdata(key), 64); err != nil {
		if bdef {
			return defval, nil
		} else {
			return defval, err
		}
	} else {
		return val, nil
	}
}

func (c *IniConfigurer) GetInt(key string, v ...int) (int, error) {
	var (
		defval int
		bdef   bool
	)

	if len(v) > 0 {
		defval = v[0]
		bdef = true
	}

	if val, err := strconv.Atoi(c.getdata(key)); err != nil {
		if bdef {
			return defval, nil
		} else {
			return defval, err
		}
	} else {
		return val, nil
	}
}

func (c *IniConfigurer) GetInt64(key string, v ...int64) (int64, error) {
	var (
		defval int64
		bdef   bool
	)

	if len(v) > 0 {
		defval = v[0]
		bdef = true
	}

	if val, err := strconv.ParseInt(c.getdata(key), 10, 64); err != nil {
		if bdef {
			return defval, nil
		} else {
			return defval, err
		}
	} else {
		return val, nil
	}
}

func (c *IniConfigurer) GetString(key string, v ...string) string {
	var defval string

	if len(v) > 0 {
		defval = v[0]
	}

	if val := c.getdata(key); val == "" {
		return defval
	} else {
		return val
	}
}

func (c *IniConfigurer) GetStrings(key string, v ...string) []string {
	var val []string
	if val = strings.Split(c.GetString(key), ","); len(val) == 1 && val[0] == "" {
		if len(v) > 0 {
			val = strings.Split(v[0], ",")
		}
	}
	return val
}

// section.key
func (c *IniConfigurer) getdata(section_key string) string {
	c.RLock()
	defer c.RUnlock()

	if len(section_key) == 0 {
		return ""
	}

	var (
		section, key string
		keys         []string = strings.Split(strings.ToLower(section_key), ".")
	)

	if len(keys) != 2 {
		panic(fmt.Sprintf("ini Set key not (session.key), error = %s", section_key))
	} else {
		section = keys[0]
		key = keys[1]
	}

	if v, ok := c.data[section]; ok {
		if vv, ok := v[key]; ok {
			return vv
		}
	}
	return ""
}

func init() {
	Register(IniProtocol, &IniConfigAdapter{})
}
