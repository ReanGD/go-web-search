package content

import (
	"database/sql/driver"
	"fmt"
	"log"
	"os"
	"reflect"
	"regexp"
	"time"

	"github.com/jinzhu/gorm"
)

// Logger implemet gorm logger interface
type Logger struct {
	gorm.LogWriter
}

var defaultLogger = Logger{log.New(os.Stdout, "\r\n", 0)}

// Format log
var sqlRegexp = regexp.MustCompile(`(\$\d+)|\?`)

// Print -
func (logger Logger) Print(values ...interface{}) {
	if len(values) > 1 {
		level := values[0]
		currentTime := "\n" + gorm.NowFunc().Format("2006-01-02 15:04:05")
		source := fmt.Sprintf("%v:", values[1])
		messages := []interface{}{source, currentTime}

		if level == "sql" {
			// duration
			messages = append(messages, fmt.Sprintf(" [%.2fms] ", float64(values[2].(time.Duration).Nanoseconds()/1e4)/100.0))
			// sql
			var formatedValues []interface{}
			for _, value := range values[4].([]interface{}) {
				indirectValue := reflect.Indirect(reflect.ValueOf(value))
				if indirectValue.IsValid() {
					value = indirectValue.Interface()
					if t, ok := value.(time.Time); ok {
						formatedValues = append(formatedValues, fmt.Sprintf("'%v'", t.Format(time.RFC3339)))
					} else if b, ok := value.([]byte); ok {
						formatedValues = append(formatedValues, fmt.Sprintf("'{...}%d'", len(b)))
						// formatedValues = append(formatedValues, fmt.Sprintf("'%v'", string(b)))
					} else if r, ok := value.(driver.Valuer); ok {
						if value, err := r.Value(); err == nil && value != nil {
							formatedValues = append(formatedValues, fmt.Sprintf("'...'"))
							// formatedValues = append(formatedValues, fmt.Sprintf("'%v'", value))
						} else {
							formatedValues = append(formatedValues, "NULL")
						}
					} else {
						formatedValues = append(formatedValues, fmt.Sprintf("'%v'", value))
					}
				} else {
					formatedValues = append(formatedValues, fmt.Sprintf("'%v'", value))
				}
			}
			messages = append(messages, fmt.Sprintf(sqlRegexp.ReplaceAllString(values[3].(string), "%v"), formatedValues...))
		} else {
			messages = append(messages, values[2:]...)
		}
		logger.Println(messages...)
	}
}
