package maillist

import (
	"fmt"
	"errors"
	"regexp"
	"strings"
	"github.com/gusluker/server-bot/src/configuration"	
)

var (
	_regLoc = regexp.MustCompile(`^[a-zA-Z\p{L} ]+(\.[a-zA-Z\p{L} ]+){0,2}$`)
	_regMail = regexp.MustCompile(`^[a-zA-Z0-9.!#$%&'*+/=?^_{|}~-]+@\w+(?:\.[a-zA-Z0-9-]+)+$`)
)

type MailList struct {
	Mails []string		
	Children map[string]*MailList
	Father *MailList
}

func Init(config *configuration.Config) (*MailList, error) {
	var world *MailList	
	var err error
	var mailList map[string]interface{}

	if mailList, err = checkSyntax(config); err == nil {
		world = New()

		for key, in := range mailList {
			key = strings.ToLower(key)
			if key == "world" {
				switch v := in.(type) {
					case []interface{}:
						for i := range v {
							mail, _ := v[i].(string)
							world.AddMail(mail)
						}
					case string:
						world.AddMail(v)
				}
			} else {
				m, _ := world.AddPath(strings.Split(key, "."))
				switch v := in.(type) {
					case []interface{}:
						for i := range v {
							mail, _ := v[i].(string)
							m.AddMail(mail)
						}
					case string:
						m.AddMail(v)
				}
			}
		}
	}

	return world, err
}

func checkSyntax(config *configuration.Config) (map[string]interface{}, error) {
	var err error
	var retval map[string]interface{}
	path := config.Paths.ConfigFilePath

	if in, ok := config.LookupOption("mail-list"); ok {
		if retval, ok = in.(map[string]interface{}); ok {
			if err = checkMailListSyntax(retval); err != nil {
				msg := fmt.Sprintf("En el archivo %s en la opción mail-list, %s", path, err)
				err = errors.New(msg)
			}
		} else {
			msg := fmt.Sprintf("En el archivo %s la opción mail-list no posee una estructura válida", path)
			err = errors.New(msg)
		}
	} else {
		err = errors.New(fmt.Sprintf("En el archivo %s no se encontró la opción mail-list", path))
	}

	return retval, err
}

func checkMailListSyntax(mailList map[string]interface{}) error {
	var err error	

	for k, v := range mailList {
		if err = checkSyntaxMailListKey(k); err != nil {
			break
		} 

		switch vm := v.(type) {
			case []interface{}:	
				for i := range vm {
					if mail, ok := vm[i].(string); ok {
						if !_regMail.MatchString(mail) {
							err = errors.New(fmt.Sprintf("El parámetro %s no posee valores válidos", k))
							break
						}
					} else {
						err = errors.New(fmt.Sprintf("El parámetro %s no posee valores válidos", k))
						break
					}
				}
			case string:
				if !_regMail.MatchString(vm) {
					err = errors.New(fmt.Sprintf("El parámetro %s no posee valores válidos", k))
				} 
			default:
				err = errors.New(fmt.Sprintf("El parámetro %s no posee valores válidos", k))
		}

		if err != nil {
			break	
		}
	}

	return err
}

func checkSyntaxMailListKey(key string) error {
	var err error

	const formatErr = "El parámetro %s no posee una estructura válida"
	if len(key) > 1 {
		low := strings.ToLower(key)
		if low != "world" {
			if !_regLoc.MatchString(low) {
				err = errors.New(fmt.Sprintf(formatErr, key))
			}
		} 
	} else {
		err = errors.New(fmt.Sprintf(formatErr, key))
	}

	return err
}

func New() *MailList {
	return &MailList {
		Father: nil,
		Children: make(map[string]*MailList),
		Mails: nil,
	}
}

func (node *MailList) AddMail(mail string) {
	node.Mails = append(node.Mails, mail)	
}

func (node *MailList) GetPath(path []string) (*MailList, bool) {
	var retval *MailList	
	ok := false 
	if isEmpty := pathIsEmpty(path); !isEmpty {
		retval = node
		for i := range path {
			if retval, ok = retval.Children[path[i]]; !ok {
				retval = nil
				break
			}
		}
	}

	return retval, ok
}

func (node *MailList) FindPath(path []string) (*MailList, bool) {
	var retval *MailList
	ok := false
	if isEmpty := pathIsEmpty(path); !isEmpty {
		retval = node
		for i := range path {
			var child *MailList
			if child, ok = retval.Children[path[i]]; !ok {
				if retval != nil && i > 0 {
					ok = true	
				}

				break
			} 

			retval = child
		}
	}

	return retval, ok
}

func pathIsEmpty(path []string) bool {
	ok := false 
	for i := range path {
		if len(path[i]) == 0 {
			ok = true
			break	
		} 
	}

	return ok
}

func (node *MailList) AddPath(path []string) (*MailList, bool) {
	var retval *MailList	
	ok := false

	void := false
	for i := range path {
		if len(path[i]) == 0 {
			void = true	
			break
		} else {
			path[i]	= strings.ToLower(path[i])
		}
	}

	if !void {
		ok = true
		ptrNode := node	
		for i := range path {
			if n, okC := ptrNode.Children[path[i]]; okC {
				ptrNode = n
			} else {
				ptrNode = ptrNode.createTree(path[i:])
				break
			}
		}

		retval = ptrNode
	}

	return retval, ok
}

func (node *MailList) createTree(path []string) *MailList {
	var retval *MailList	
	if path != nil {
		n := node
		for i := range path {
			n, _ = n.AppendChild(path[i])
		}

		retval = n
	}

	return retval
}

func (node *MailList) AppendChild(name string) (*MailList, bool) {
	var child *MailList
	okAdd := false

	if _, ok := node.Children[name]; !ok {
		okAdd = true
		child = New()
		child.Father = node
		node.Children[name] = child
	}

	return child, okAdd
}

func (node *MailList) GetMailsHierarchically() ([]string, bool) {
	retval := node
	ok := false

	for {
		if retval.Mails != nil {
			ok = true
			break
		} else if retval.Father != nil {
			retval = retval.Father
			continue
		} else {
			ok = false
			break
		}
	}

	return retval.Mails, ok
}

func (node *MailList) String() string {
	retval := "Arbol vacío"

	if node != nil {
		if len(node.Children) > 0 {
			retval = node.printNode("")
		}
	}

	return retval
}

func (node *MailList) printNode(msg string) string {
	var retval string	

	if len(node.Children) > 0 {
		for k, v := range node.Children {
			m := msg + fmt.Sprintf("->%s", k)
			retval += v.printNode(m) 
		}
	} else {
		retval = msg + "\n"
	}

	return retval
}
