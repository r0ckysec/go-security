package goflags

import (
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/cnf/structhash"
)

// FlagSet is a list of flags for an application
type FlagSet struct {
	Marshal     bool
	description string
	flagKeys    InsertionOrderedMap
	groups      []groupData
	CommandLine *flag.FlagSet

	// OtherOptionsGroupName is the name for all flags not in a group
	OtherOptionsGroupName string
}

type groupData struct {
	name        string
	description string
}

type FlagData struct {
	usage        string
	short        string
	long         string
	group        string // unused unless set later
	defaultValue interface{}
	skipMarshal  bool
}

// Group sets the group for a flag data
func (flagData *FlagData) Group(name string) {
	flagData.group = name
}

// NewFlagSet creates a new flagSet structure for the application
func NewFlagSet() *FlagSet {
	return &FlagSet{flagKeys: *newInsertionOrderedMap(), OtherOptionsGroupName: "other options", CommandLine: flag.NewFlagSet(os.Args[0], flag.ExitOnError)}
}

func newInsertionOrderedMap() *InsertionOrderedMap {
	return &InsertionOrderedMap{
		values: make(map[string]*FlagData),
		keys:   make([]string, 0),
	}
}

// Hash returns the unique hash for a flagData structure
// NOTE: Hash panics when the structure cannot be hashed.
func (flagData *FlagData) Hash() string {
	hash, _ := structhash.Hash(flagData, 1)
	return hash
}

// SetDescription sets the description field for a flagSet to a value.
func (flagSet *FlagSet) SetDescription(description string) {
	flagSet.description = description
}

// SetGroup sets a group with name and description for the command line options
//
// The order in which groups are passed is also kept as is, similar to flags.
func (flagSet *FlagSet) SetGroup(name, description string) {
	flagSet.groups = append(flagSet.groups, groupData{name: name, description: description})
}

func (flagSet *FlagSet) CreateGroup(groupName, description string, flags ...*FlagData) {
	flagSet.SetGroup(groupName, description)
	for _, currentFlag := range flags {
		currentFlag.Group(groupName)
	}
}

// Parse parses the flags provided to the library.
func (flagSet *FlagSet) Parse() error {
	flagSet.CommandLine.Usage = flagSet.usageFunc
	_ = flagSet.CommandLine.Parse(os.Args[1:])
	return nil
}

// VarP adds a Var flag with a shortname and longname
func (flagSet *FlagSet) VarP(field flag.Value, long, short, usage string) *FlagData {
	flagSet.CommandLine.Var(field, short, usage)
	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: field,
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// Var adds a Var flag with a longname
func (flagSet *FlagSet) Var(field flag.Value, long, usage string) *FlagData {
	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: field,
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// StringVarEnv adds a string flag with a shortname and longname with a default value read from env variable
// with a default value fallback
func (flagSet *FlagSet) StringVarEnv(field *string, long, short, defaultValue, envName, usage string) *FlagData {
	if envValue, exists := os.LookupEnv(envName); exists {
		defaultValue = envValue
	}
	return flagSet.StringVarP(field, long, short, defaultValue, usage)
}

// StringVarP adds a string flag with a shortname and longname
func (flagSet *FlagSet) StringVarP(field *string, long, short, defaultValue, usage string) *FlagData {
	flagSet.CommandLine.StringVar(field, short, defaultValue, usage)
	flagSet.CommandLine.StringVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// StringVar adds a string flag with a longname
func (flagSet *FlagSet) StringVar(field *string, long, defaultValue, usage string) *FlagData {
	flagSet.CommandLine.StringVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// BoolVarP adds a bool flag with a shortname and longname
func (flagSet *FlagSet) BoolVarP(field *bool, long, short string, defaultValue bool, usage string) *FlagData {
	flagSet.CommandLine.BoolVar(field, short, defaultValue, usage)
	flagSet.CommandLine.BoolVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: strconv.FormatBool(defaultValue),
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// BoolVar adds a bool flag with a longname
func (flagSet *FlagSet) BoolVar(field *bool, long string, defaultValue bool, usage string) *FlagData {
	flagSet.CommandLine.BoolVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: strconv.FormatBool(defaultValue),
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// IntVarP adds a int flag with a shortname and longname
func (flagSet *FlagSet) IntVarP(field *int, long, short string, defaultValue int, usage string) *FlagData {
	flagSet.CommandLine.IntVar(field, short, defaultValue, usage)
	flagSet.CommandLine.IntVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: strconv.Itoa(defaultValue),
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// IntVar adds a int flag with a longname
func (flagSet *FlagSet) IntVar(field *int, long string, defaultValue int, usage string) *FlagData {
	flagSet.CommandLine.IntVar(field, long, defaultValue, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: strconv.Itoa(defaultValue),
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// NormalizedStringSliceVarP adds a path slice flag with a shortname and longname.
// It supports comma separated values, that are normalized (lower-cased, stripped of any leading and trailing whitespaces and quotes)
func (flagSet *FlagSet) NormalizedStringSliceVarP(field *NormalizedStringSlice, long, short string, defaultValue NormalizedStringSlice, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, short, usage)
	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// NormalizedStringSliceVar adds a path slice flag with a long name
// It supports comma separated values, that are normalized (lower-cased, stripped of any leading and trailing whitespaces and quotes)
func (flagSet *FlagSet) NormalizedStringSliceVar(field *NormalizedStringSlice, long string, defaultValue NormalizedStringSlice, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// StringSliceVarP adds a string slice flag with a shortname and longname
// Supports ONE value at a time. Adding multiple values require repeating the argument (-flag value1 -flag value2)
// No value normalization is happening.
func (flagSet *FlagSet) StringSliceVarP(field *StringSlice, long, short string, defaultValue StringSlice, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, short, usage)
	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// StringSliceVar adds a string slice flag with a longname
// Supports ONE value at a time. Adding multiple values require repeating the argument (-flag value1 -flag value2)
// No value normalization is happening.
func (flagSet *FlagSet) StringSliceVar(field *StringSlice, long string, defaultValue StringSlice, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: defaultValue,
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// RuntimeMapVarP adds a runtime only map flag with a longname
func (flagSet *FlagSet) RuntimeMapVar(field *RuntimeMap, long string, defaultValue []string, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		long:         long,
		defaultValue: defaultValue,
		skipMarshal:  true,
	}
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

// RuntimeMapVarP adds a runtime only map flag with a shortname and longname
func (flagSet *FlagSet) RuntimeMapVarP(field *RuntimeMap, long, short string, defaultValue []string, usage string) *FlagData {
	for _, item := range defaultValue {
		_ = field.Set(item)
	}

	flagSet.CommandLine.Var(field, short, usage)
	flagSet.CommandLine.Var(field, long, usage)

	flagData := &FlagData{
		usage:        usage,
		short:        short,
		long:         long,
		defaultValue: defaultValue,
		skipMarshal:  true,
	}
	flagSet.flagKeys.Set(short, flagData)
	flagSet.flagKeys.Set(long, flagData)
	return flagData
}

func (flagSet *FlagSet) usageFunc() {
	cliOutput := flagSet.CommandLine.Output()
	fmt.Fprintf(cliOutput, "%s\n\n", flagSet.description)
	fmt.Fprintf(cliOutput, "Usage:\n  %s [flags]\n\n", os.Args[0])
	fmt.Fprintf(cliOutput, "Flags:\n")

	writer := tabwriter.NewWriter(cliOutput, 0, 0, 1, ' ', 0)

	if len(flagSet.groups) > 0 {
		flagSet.usageFuncForGroups(cliOutput, writer)
	} else {
		flagSet.usageFuncInternal(writer)
	}
}

// usageFuncInternal prints usage for command line flags
func (flagSet *FlagSet) usageFuncInternal(writer *tabwriter.Writer) {
	uniqueDeduper := newUniqueDeduper()

	flagSet.flagKeys.forEach(func(key string, data *FlagData) {
		currentFlag := flagSet.CommandLine.Lookup(key)

		if !uniqueDeduper.isUnique(data) {
			return
		}
		result := createUsageString(data, currentFlag)
		fmt.Fprint(writer, result, "\n")
	})
	writer.Flush()
}

// usageFuncForGroups prints usage for command line flags with grouping enabled
func (flagSet *FlagSet) usageFuncForGroups(cliOutput io.Writer, writer *tabwriter.Writer) {
	uniqueDeduper := newUniqueDeduper()

	var otherOptions []string
	for _, group := range flagSet.groups {
		fmt.Fprintf(cliOutput, "%s:\n", normalizeGroupDescription(group.description))

		flagSet.flagKeys.forEach(func(key string, data *FlagData) {
			currentFlag := flagSet.CommandLine.Lookup(key)

			if data.group == "" {
				if !uniqueDeduper.isUnique(data) {
					return
				}
				otherOptions = append(otherOptions, createUsageString(data, currentFlag))
				return
			}
			// Ignore the flag if it's not in our intended group
			if !strings.EqualFold(data.group, group.name) {
				return
			}
			if !uniqueDeduper.isUnique(data) {
				return
			}
			result := createUsageString(data, currentFlag)
			fmt.Fprint(writer, result, "\n")
		})
		writer.Flush()
		fmt.Printf("\n")
	}

	// Print Any additional flag that may have been left
	if len(otherOptions) > 0 {
		fmt.Fprintf(cliOutput, "%s:\n", normalizeGroupDescription(flagSet.OtherOptionsGroupName))

		for _, option := range otherOptions {
			fmt.Fprint(writer, option, "\n")
		}
		writer.Flush()
	}
}

type uniqueDeduper struct {
	hashes map[string]interface{}
}

func newUniqueDeduper() *uniqueDeduper {
	return &uniqueDeduper{hashes: make(map[string]interface{})}
}

// isUnique returns true if the flag is unique during iteration
func (u *uniqueDeduper) isUnique(data *FlagData) bool {
	dataHash := data.Hash()
	if _, ok := u.hashes[dataHash]; ok {
		return false // Don't print the value if printed previously
	}
	u.hashes[dataHash] = struct{}{}
	return true
}

func isNotBlank(value string) bool {
	return len(strings.TrimSpace(value)) != 0
}

func createUsageString(data *FlagData, currentFlag *flag.Flag) string {
	valueType := reflect.TypeOf(currentFlag.Value)

	result := createUsageFlagNames(data)
	result += createUsageTypeAndDescription(currentFlag, valueType)
	result += createUsageDefaultValue(data, currentFlag, valueType)

	return result
}

func createUsageDefaultValue(data *FlagData, currentFlag *flag.Flag, valueType reflect.Type) string {
	if !isZeroValue(currentFlag, currentFlag.DefValue) {
		defaultValueTemplate := " (default "
		switch valueType.String() { // ugly hack because "flag.stringValue" is not exported from the parent library
		case "*flag.stringValue":
			defaultValueTemplate += "%q"
		default:
			defaultValueTemplate += "%v"
		}
		defaultValueTemplate += ")"
		return fmt.Sprintf(defaultValueTemplate, data.defaultValue)
	}
	return ""
}

func createUsageTypeAndDescription(currentFlag *flag.Flag, valueType reflect.Type) string {
	var result string

	flagDisplayType, usage := flag.UnquoteUsage(currentFlag)
	if len(flagDisplayType) > 0 {
		if flagDisplayType == "value" { // hardcoded in the goflags library
			switch valueType.Kind() {
			case reflect.Ptr:
				pointerTypeElement := valueType.Elem()
				switch pointerTypeElement.Kind() {
				case reflect.Slice, reflect.Array:
					switch pointerTypeElement.Elem().Kind() {
					case reflect.String:
						flagDisplayType = "string[]"
					default:
						flagDisplayType = "value[]"
					}
				}
			}
		}
		result += " " + flagDisplayType
	}

	result += "\t\t"
	result += strings.ReplaceAll(usage, "\n", "\n"+strings.Repeat(" ", 4)+"\t")
	return result
}

func createUsageFlagNames(data *FlagData) string {
	flagNames := strings.Repeat(" ", 2) + "\t"

	var validFlags []string
	addValidParam := func(value string) {
		if isNotBlank(value) {
			validFlags = append(validFlags, fmt.Sprintf("-%s", value))
		}
	}

	addValidParam(data.short)
	addValidParam(data.long)

	if len(validFlags) == 0 {
		panic("CLI arguments cannot be empty.")
	}

	flagNames += strings.Join(validFlags, ", ")
	return flagNames
}

// isZeroValue determines whether the string represents the zero
// value for a flag.
func isZeroValue(f *flag.Flag, value string) bool {
	// Build a zero value of the flag's Value type, and see if the
	// result of calling its String method equals the value passed in.
	// This works unless the Value type is itself an interface type.
	valueType := reflect.TypeOf(f.Value)
	var zeroValue reflect.Value
	if valueType.Kind() == reflect.Ptr {
		zeroValue = reflect.New(valueType.Elem())
	} else {
		zeroValue = reflect.Zero(valueType)
	}
	return value == zeroValue.Interface().(flag.Value).String()
}

// normalizeGroupDescription returns normalized description field for group
func normalizeGroupDescription(description string) string {
	return strings.ToUpper(description)
}
