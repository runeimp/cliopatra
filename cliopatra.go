package cliopatra

import (
	"errors"
	"fmt"
	"log"
	"math/bits"
	"os"
	"strconv"
	"strings"
)

/*
 * CONSTANTS
 */
const (
	DefaultPrefix           = "-"
	DefaultSuffix           = ""
	ErrorArgumentMissing    = "argument not set"
	ErrorFlagMissing        = "the flag was not set"
	ErrorKeyLengthZero      = "the key length must be greater than zero"
	ErrorNameLengthZero     = "the parameter name length must be greater than zero"
	ErrorOptionMissing      = "the option was not set"
	ErrorOptionValueMissing = "the option's value was not set"
	ErrorEnvDefaultSetEmpty = "the environment variable default name must not be empty (zero length or all whitespace)"
	PackageVersion          = "0.1.0-alpha"
	StringTruthyOne         = "1"
	StringTruthyTrue        = "true"
	StringTruthyYes         = "yes"
)

/*
 * DERIVED CONSTANTS
 */
var (
	cliopatraInstance *Cliopatra // Singleton
	intSize           = bits.UintSize
)

/*
 * TYPES
 */

// Cliopatra is the extended root CommandSet
type Cliopatra struct {
	*CommandSet
	CliApp string
}

// // GetHelp returns the help info for the Cliopatra instance
// func (c *Cliopatra) GetHelp() string {
// 	help := c.CommandSet.Name
// 	for pk, pv := range c.Parameters {
// 		help += fmt.Sprintf(`
// USAGE: %q

// OPTIONS:
// `, c.CliApp)
// 		// help += fmt.Sprintf("  %-20s  %s\n", pk, pv.GetHelp())
// 		// help += fmt.Sprintf("  %-20s  %q\n", pk, pv.Parameter.Name)
// 		prefixes := ""
// 		for _, v := range pv.GetName() {
// 			// fmt.Printf("  %s  %q\n", pk, v)
// 			for _, p := range pv.GetPrefix() {
// 				// fmt.Printf("  %s    %q\n", pk, p)
// 				prefixes += " | " + p + v
// 			}
// 		}
// 		help += fmt.Sprintf("  %-20s  %s\n", prefixes[3:], pv.GetHelp()+" ("+pk+")")
// 	}
// 	return help
// }

// Run processes the command line parameters
func (c *Cliopatra) Run() {
	c.CliApp = os.Args[0]
	for i, v := range os.Args[1:] {
		log.Printf("%0.3d  %q\n", i, v)
		matchList = append(matchList, matchItem{
			Index:   i + 1,
			Matched: false,
			Value:   v,
		})
	}
	fmt.Println()
	c.CommandSet.MatchCommandLine(matchList)
}

type matchItem struct {
	Index   int
	Matched bool
	Value   string
}

// CommandLineParameter is the data interface for command line parameters
type CommandLineParameter interface {
	GetFlag() bool               // Returns the boolean value of a flag
	GetHelp() string             // Returns the help info for the parameter
	GetInt() (int, error)        // Returns the value as a system integer
	GetName() []string           // Returns the parameters names available on the command line
	GetNumber() (float64, error) // Returns the value as a float64
	GetPrefix() []string         // Returns the valid values for to prefix the parameter names
	GetUint() (uint, error)      // Returns the value as a system unsigned integer
	GetValue() (string, error)   // Returns the value as a string
	SetDefault(string) error     // Defines the default value to use if there is no value given on the command line and no environment variable default found
	SetFlag()                    // Defined the flag was used on the command line
	SetKey(string) error         // Defines the key name to reference in code
	SetName([]string) error      // Defines the parameter name(s) allowed on the command line
	SetValue(string)             // Defines the parameter value found on the command line
	SetPrefix([]string, bool)    // Defines allowed alternate or custom prefixes to be used instead of, or in addition to, the command set prefix(s). Default prefix is a hyphen.
	SetRequired(bool)            // Defines the parameter as required input. Errors if not present on the command line.
}

// CommandSet is the data type for command line subcommand
type CommandSet struct {
	AllowPosixGroups bool   // Allow POSIX option groups?
	Description      string // The long description to display to the user
	Help             string // The help information to display to the user
	IsGNU            bool   // Does the parameter conform to the GNU specification
	IsMultics        bool   // Does the parameter conform to the Multics specification
	IsPosix          bool   // Does the parameter conform to the POSIX specification
	IsRuneImp        bool   // Does the parameter conform to the RuneImp specification
	Name             string // The name of the command set
	Parameters       map[string]CommandLineParameter
	Prefix           []string // List of allowed parameter prefixes. Mostly used for options/flags. Though occasionally used for arguments.
	Suffix           []string // List of allowed parameter suffixes. Mostly used for arguments. Though occasionally used for options/flags.
	Summery          string   // The short description to display to the user
}

// AddFlag defines a flag parameter for a command set
func (cs CommandSet) AddFlag(key string, name []string, prefix *[]string, help string) {
	p := cs.Prefix
	if prefix != nil && len(*prefix) > 0 {
		p = *prefix
	}

	cs.Parameters[key] = &Flag{
		Parameter: Parameter{
			help:   help,
			Name:   name,
			Prefix: p,
		},
		defaultValue: false,
	}
}

// GetHelp returns the help info for the command set
func (cs CommandSet) GetHelp() string {
	help := cs.Name
	help += fmt.Sprintf(`

OPTIONS:
`)
	for pk, pv := range cs.Parameters {
		// help += fmt.Sprintf("  %-20s  %s\n", pk, pv.GetHelp())
		// help += fmt.Sprintf("  %-20s  %q\n", pk, pv.Parameter.Name)
		prefixes := ""
		for _, v := range pv.GetName() {
			// fmt.Printf("  %s  %q\n", pk, v)
			for _, p := range pv.GetPrefix() {
				// fmt.Printf("  %s    %q\n", pk, p)
				prefixes += " | " + p + v
			}
		}
		help += fmt.Sprintf("  %-20s  %s\n", prefixes[3:], pv.GetHelp()+" ("+pk+")")
	}
	return help
}

// MatchCommandLine helps process matches for each command set
func (cs CommandSet) MatchCommandLine(args []matchItem) bool {

	for i, cl := range args {
	ArgsContinue:
		for pk, pv := range cs.Parameters {
			for _, v := range pv.GetName() {
				for _, p := range pv.GetPrefix() {
					// NOTE: Need better testing. Need to check for --option=value with strings.StartsWith, etc.
					if cl.Value == p+v {
						fmt.Printf("%s | %s | %s%s | ***MATCH***\n", cl.Value, pk, p, v)
						args[i].Matched = true
						switch t := pv.(type) {
						case *Flag:
							pv.SetFlag()
							// cs.Parameters[pk].SetFlag()
						case *Argument:
							// NOTE: Fixed position or trailing?
							pv.SetValue(cl.Value)
						case *Option:
							// NOTE: Required value? How should I handle the option value testing?
							pv.SetValue(cl.Value)
						default:
							log.Printf("type = %T\n", t)
						}
						break ArgsContinue
					} else {
						fmt.Printf("%s | %s | %s%s\n", cl.Value, pk, p, v)
					}
				}
			}
		}
	}

	for i, cl := range args {
		log.Printf("Args | %02d | %#v\n", i, cl)
	}

	return false
}

// SetGNU defines if parameter can use GNU long option names.
// Word based names with a double hyphen prefix and potential hyphen word separation.
func (cs *CommandSet) SetGNU(v bool) {
	cs.IsGNU = v
}

// SetMultics defines if the parameter can use Multics option names.
// One or more letter or word names with a single hyphen prefix and potential hyphen or underscore based word separation.
func (cs *CommandSet) SetMultics(v bool) {
	cs.IsMultics = v
}

// SetPosix defines if the parameter can use POSIX short option names.
// A single letter name with a single hyphen prefix.
func (cs *CommandSet) SetPosix(v bool) {
	cs.IsPosix = v
}

// SetPosixGroups defines if POSIX group processing is done. i.e.: a group of letters prefixed with a single hyphen are expanded into individual POSIX options.
// Should not be combined with Multics options.
func (cs *CommandSet) SetPosixGroups(v bool) {
	cs.AllowPosixGroups = v
}

// SetRuneImp defines if the parameter can use RuneImp option names.
// Multics plus Grouping of multiple single letter names prefixed with a double hyphen and no name separation.
// Can not be combined with POSIX groups.
func (cs *CommandSet) SetRuneImp(v bool) {
	cs.IsRuneImp = v
}

// Argument is the data type for command line arguments
type Argument struct {
	Parameter
	defaultValue string // The default value to use if one is not given on the command line
}

// GetFlag returns the current boolean value for the parameter
func (a *Argument) GetFlag() bool {
	v, err := a.GetValue()
	if err != nil {
		return false
	}
	return truthyString(v)
}

// GetHelp returns the help info for the parameter
func (a *Argument) GetHelp() string {
	return a.Parameter.help
}

// GetInt returns the value as a system integer
func (a *Argument) GetInt() (int, error) {
	i, err := strconv.Atoi(a.Parameter.value)
	return i, err
}

// GetName returns the parameters names available on the command line
func (a *Argument) GetName() []string {
	return []string{}
}

// GetNumber returns the value as a float64
func (a *Argument) GetNumber() (float64, error) {
	i, err := strconv.ParseFloat(a.Parameter.value, 64)
	return i, err
}

// GetPrefix returns the valid values for to prefix the parameter names
func (a *Argument) GetPrefix() []string {
	return a.Parameter.Prefix
}

// GetUint returns the value as a system unsigned integer
func (a *Argument) GetUint() (uint, error) {
	i, err := strconv.ParseUint(a.Parameter.value, 10, intSize)
	// ui := uint(i)
	return uint(i), err
}

// GetValue returns the current value for the parameter
func (a *Argument) GetValue() (string, error) {
	if a.valueSet == false {
		if a.defaultSet {
			if a.configPreferred && len(a.configDefault) > 0 {
				a.value = a.configDefault
			} else if len(a.envDefault) > 0 {
				a.value = a.envDefault
			} else {
				a.value = a.configDefault
			}
		} else {
			return "", errors.New(ErrorArgumentMissing)
		}
	}
	return a.value, nil
}

// SetDefault defines the default value to use if there is no value given on the command line and no environment or config variable default found
func (a *Argument) SetDefault(s string) error {
	a.value = s
	return nil
}

// SetFlag is a NO-OP that will panic for this parameter type
func (a *Argument) SetFlag() {
	panic("SetFlag use not appropriate for this parameter type")
}

// SetKey defines the key name to reference in code
func (a *Argument) SetKey(s string) error {
	if len(s) == 0 {
		return errors.New(ErrorKeyLengthZero)
	}
	a.Key = s
	return nil
}

// SetName defines the parameter name(s) allowed on the command line
func (a *Argument) SetName(s []string) error {
	for _, v := range s {
		name := strings.TrimSpace(v)
		if len(name) == 0 {
			return errors.New(ErrorNameLengthZero)
		}
	}

	a.Name = s
	return nil
}

// SetPrefix allows for alternate or custom prefixes to be used. Default prefix is a hyphen.
func (a *Argument) SetPrefix(list []string, appendToList bool) {
	if appendToList {
		for _, v := range list {
			a.Prefix = append(a.Prefix, v)
		}
	} else {
		a.Prefix = list
	}
}

// SetRequired defines the parameter as required input. Errors if not present on the command line.
func (a *Argument) SetRequired(b bool) {
	a.IsRequired = b
}

// SetValue defines the command line value given
func (a *Argument) SetValue(s string) {
	a.value = s
	a.valueSet = true
}

// Flag is the data type for command line flags
type Flag struct {
	Parameter
	defaultValue bool // The default value to use if one is not given on the command line. Default: false
	flagValue    bool // The actual value of the parameter given
}

// GetFlag returns the current boolean value for the parameter
func (f *Flag) GetFlag() bool {
	return f.flagValue
}

// GetHelp returns the help info for the parameter
func (f *Flag) GetHelp() string {
	return f.Parameter.help
}

// GetInt returns the value as a system integer
func (f *Flag) GetInt() (int, error) {
	if f.flagValue {
		return 1, nil
	}
	return 0, nil
}

// GetName returns the parameters names available on the command line
func (f *Flag) GetName() []string {
	return f.Parameter.Name
}

// GetNumber returns the value as a float64
func (f *Flag) GetNumber() (float64, error) {
	if f.flagValue {
		return 1.0, nil
	}
	return 0.0, nil
}

// GetPrefix returns the valid values for to prefix the parameter names
func (f *Flag) GetPrefix() []string {
	return f.Parameter.Prefix
}

// GetUint returns the value as a system unsigned integer
func (f *Flag) GetUint() (uint, error) {
	if f.flagValue {
		return 1, nil
	}
	return 0, nil
}

// GetValue returns the current value for the parameter
func (f *Flag) GetValue() (string, error) {
	// if f.valueSet == false {
	// 	if f.defaultSet {
	// 		if f.configPreferred && len(f.configDefault) > 0 {
	// 			if truthyString(f.configDefault) {
	// 				f.flagValue = true
	// 			}
	// 		} else if len(f.envDefault) > 0 {
	// 			f.flagValue = truthyString(f.envDefault)
	// 		} else {
	// 			f.flagValue = truthyString(f.configDefault)
	// 		}
	// 	}
	// }
	if f.flagValue {
		return "true", nil
	}
	return "false", nil
}

// SetDefault defines the default value to use if there is no value given on the command line and no environment or config variable default found
func (f *Flag) SetDefault(s string) error {
	f.value = s
	f.flagValue = truthyString(s)
	return nil
}

// SetFlag defines the command line flag as set
func (f *Flag) SetFlag() {
	log.Printf("Flag.SetFlag() | true\n")
	f.flagValue = true
	f.Parameter.value = "true"
	f.Parameter.valueSet = true
	log.Printf("Flag.SetFlag() | f.flagValue = %t\n", f.flagValue)
	log.Printf("Flag.SetFlag() | f.GetFlag() = %t\n", f.GetFlag())
}

// SetKey defines the key name to reference in code
func (f *Flag) SetKey(s string) error {
	if len(s) == 0 {
		return errors.New(ErrorKeyLengthZero)
	}
	f.Key = s
	return nil
}

// SetName defines the parameter name(s) allowed on the command line
func (f *Flag) SetName(s []string) error {
	for _, v := range s {
		name := strings.TrimSpace(v)
		if len(name) == 0 {
			return errors.New(ErrorNameLengthZero)
		}
	}

	f.Name = s
	return nil
}

// SetPrefix allows for alternate or custom prefixes to be used. Default prefix is a hyphen.
func (f *Flag) SetPrefix(list []string, appendToList bool) {
	if appendToList {
		for _, v := range list {
			f.Prefix = append(f.Prefix, v)
		}
	} else {
		f.Prefix = list
	}
}

// SetRequired defines the parameter as required input. Errors if not present on the command line.
func (f *Flag) SetRequired(b bool) {
	f.IsRequired = b
}

// SetValue defines the command line value given
func (f *Flag) SetValue(s string) {
	log.Printf("Flag.SetValue() | %q\n", s)
	f.flagValue = truthyString(s)
	f.value = s
	f.valueSet = true
}

// Option is the data type for command line options
type Option struct {
	Parameter
	defaultValue string // The default value to use if one is not given on the command line
}

// GetFlag returns the current boolean value for the parameter
func (o *Option) GetFlag() bool {
	s, err := o.GetValue()
	if err != nil {
		return false
	}
	return truthyString(s)
}

// GetHelp returns the help info for the parameter
func (o *Option) GetHelp() string {
	return o.Parameter.help
}

// GetInt returns the value as a system integer
func (o *Option) GetInt() (int, error) {
	i, err := strconv.Atoi(o.Parameter.value)
	return i, err
}

// GetName returns the parameters names available on the command line
func (o *Option) GetName() []string {
	return o.Parameter.Name
}

// GetNumber returns the value as a float64
func (o *Option) GetNumber() (float64, error) {
	i, err := strconv.ParseFloat(o.Parameter.value, 64)
	return i, err
}

// GetPrefix returns the valid values for to prefix the parameter names
func (o *Option) GetPrefix() []string {
	return o.Parameter.Prefix
}

// GetUint returns the value as a system unsigned integer
func (o *Option) GetUint() (uint, error) {
	i, err := strconv.ParseUint(o.Parameter.value, 10, intSize)
	// ui := uint(i)
	return uint(i), err
}

// GetValue returns the current value for the parameter
func (o *Option) GetValue() (string, error) {
	if o.valueSet == false {
		if o.defaultSet {
			if o.configPreferred && len(o.configDefault) > 0 {
				o.value = o.configDefault
			} else if len(o.envDefault) > 0 {
				o.value = o.envDefault
			} else {
				o.value = o.configDefault
			}
		} else {
			return "", errors.New(ErrorOptionMissing)
		}
	}
	return o.value, nil
}

// SetDefault defines the default value to use if there is no value given on the command line and no environment or config variable default found
func (o *Option) SetDefault(s string) error {
	o.value = s
	return nil
}

// SetFlag is a NO-OP that will panic for this parameter type
func (o *Option) SetFlag() {
	panic("SetFlag use not appropriate for this parameter type")
}

// SetKey defines the key name to reference in code
func (o *Option) SetKey(s string) error {
	if len(s) == 0 {
		return errors.New(ErrorKeyLengthZero)
	}
	o.Key = s
	return nil
}

// SetPrefix allows for alternate or custom prefixes to be used. Default prefix is a hyphen.
func (o *Option) SetPrefix(list []string, appendToList bool) {
	if appendToList {
		for _, v := range list {
			o.Prefix = append(o.Prefix, v)
		}
	} else {
		o.Prefix = list
	}
}

// SetName defines the parameter name(s) allowed on the command line
func (o *Option) SetName(s []string) error {
	for _, v := range s {
		name := strings.TrimSpace(v)
		if len(name) == 0 {
			return errors.New(ErrorNameLengthZero)
		}
	}
	o.Name = s
	return nil
}

// SetValue defines the command line value given
func (o *Option) SetValue(s string) {
	o.value = s
	o.valueSet = true
}

// SetRequired defines the parameter as required input. Errors if not present on the command line.
func (o *Option) SetRequired(b bool) {
	o.IsRequired = b
}

// Parameter is the data type for all options and arguments
type Parameter struct {
	configDefault   string   // Configuration variable to use as a default value if the parameter is not present on the command line
	configPreferred bool     // If the external default preference should be for the config file over an environment variables. Default: false
	defaultSet      bool     // If there is a default value to check for
	Description     string   // The long description to display to the user
	envDefault      string   // Environment variable to use as a default value if the parameter is not present on the command line
	help            string   // The help information to display to the user
	Index           int      // The actual index on the command line. Default: 0 (equates to "not set" as the zeroth position is the command itself)
	IsRequired      bool     // Defines if this parameter is required on the command line
	Key             string   // The key used in the Parameters map
	MetaVar         string   // The variable reference name used in example usage
	Name            []string // The command line name(s) allowed
	Position        uint     // Is the arguments position fixed. Useful for subcommands and many tools. Default: 0 (position not fixed)
	Prefix          []string // List of allowed parameter prefixes. Mostly used for options/flags. Though occasionally used for arguments.
	Suffix          []string // List of allowed parameter suffixes. Mostly used for arguments. Though occasionally used for options/flags.
	Summery         string   // The short description to display to the user
	value           string   // The actual value of the parameter given
	valueRequired   bool     // Defines if this parameter's value is required when the parameter is used
	valueSet        bool     // The flag was set or actual value of the parameter was given
}

/*
 * VARIABLES
 */
var (
	matchList = []matchItem{}
)

/*
 * FUNCTIONS
 */

// New returns a CLIOPATra singleton instance
func New(cs CommandSet) (*Cliopatra, error) {
	if cliopatraInstance != nil {
		return cliopatraInstance, nil
	}

	cs.Parameters = make(map[string]CommandLineParameter)
	if len(cs.Prefix) == 0 && len(DefaultPrefix) > 0 {
		cs.Prefix = []string{DefaultPrefix}
	}
	if len(cs.Suffix) == 0 && len(DefaultSuffix) > 0 {
		cs.Suffix = []string{DefaultSuffix}
	}

	cliopatraInstance := &Cliopatra{CommandSet: &cs}

	return cliopatraInstance, nil
}

func truthyString(s string) bool {
	str := strings.ToLower(s)
	switch str {
	case StringTruthyOne, StringTruthyTrue, StringTruthyYes:
		return true
	}
	return false
}
