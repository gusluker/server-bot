package maillist

import (
	"os"
	"strings"
	"testing"

	"github.com/gusluker/server-bot/server/configuration"
)

type TestMailListType struct {
	config *configuration.Config
	path string
	tree map[string]interface{}
	size map[string]int
	res *MailList
	err bool
}

var (
	pathMails = "github.com/gusluker/server-bot/test/emails1.sbot"

	testTable = []TestMailListType{
		{
			config: nil	,
			path: "github.com/gusluker/server-bot/test/config3.sbot",
			tree: map[string]interface{} {
				"world": []interface{}{
					"correo1@correo.com", 
					"correo2@correo.com",
				},
				"co.valle del cauca.cali": []interface{} {
					"correo3@correo.com", 
					"correo4@correo.com",
					"correo5@correo.com",
				},
				"co.valle del cauca.buenaventura": "correo6@correo.com",
				"co.huila.pitalito": "correo7@correo.com",
				"co.huila.palestina": "correo8@correo.com",
				"co.valle del cauca": "correo9@correo.com",
				"co.huila": "correo10@correo.com",
				"co": "correo11@correo.com",
				"ec.pichincha.quito": "correo1@proton.com",
				"ec.manabí.manta": []interface{}{
					"correo2@proton.com",
					"correo3@proton.com",
				},
				"ec.manabí.jaramijó":"correo4@proton.com",
				"ec.manabí": []interface{} {
					"correo5@proton.com",
					"correo6@proton.com",
				},
				"ec":"correo7@proton.com",
			},
			size: map[string]int {
				"world": 2,
				"co.valle del cauca.cali": 0,
				"co.valle del cauca.buenaventura": 0,
				"co.huila.pitalito": 0,
				"co.huila.palestina": 0,
				"co.valle del cauca": 2,
				"co.huila": 2,
				"co": 2,
				"ec.pichincha.quito": 0,
				"ec.manabí.manta": 0,
				"ec.manabí.jaramijó": 0,
				"ec.manabí": 2,
				"ec": 2,
			},
			err: false,
		},
	}
)

func TestAddPath(t *testing.T) {
	for i := range testTable {
		world := New()	
		path := strings.Split("padre1..nieto2", ".")
		if _, ok := world.AddPath(path); ok {
			t.Fatal("No debió crearse el árbol")
		}

		path = strings.Split(".hijo1.nieto2", ".")
		if _, ok := world.AddPath(path); ok {
			t.Fatal("No debió crearse el árbol")
		}

		path = strings.Split("padre1.hijo1.", ".")
		if _, ok := world.AddPath(path); ok {
			t.Fatal("No debió crearse el árbol")
		}

		world = createTreeTest(testTable[i].tree, t)
		t.Logf("Tree: \n%s", world)

		for k, v := range testTable[i].size {
			if k == "world"	{
				s, _ := testTable[i].size["world"]
				if len(world.Children) != s {
					t.Fatalf("Ciclo %d. No concuerda los hijos. Esperados %d == %d Recibidos", i, s, len(world.Children))	
				}

				continue
			}

			var node *MailList
			ok := false
			if node, ok = world.GetPath(strings.Split(k, ".")); !ok {
				t.Fatalf("Ciclo %d. No existe la ruta %s\nArbol: %s", i, k, node)
			}

			if len(node.Children) != v {
				t.Fatalf("En ruta %s no concuerdan sus hijos. Esperados %d == %d Recibidos", k, v, len(node.Children))
			}
		}
	}
}

func createTreeTest(tree map[string]interface{}, t *testing.T) *MailList {
	world := New()	

	for key, in := range tree {
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
			var m *MailList
			ok := false
			if m, ok = world.AddPath(strings.Split(key, ".")); !ok || m == nil {
				t.Fatalf("La ruta debe ser válida")	
			}

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

	return world
}

func TestInit(t *testing.T) {
	initTest(t)

	for i := range testTable {
		world := createTreeTest(testTable[i].tree, t)

		if len(testTable[i].res.Children) != len(world.Children) {
			t.Fatal("No concuerdan el número de hijos de world.")
		}
		
		cmpTree("world", testTable[i].res, world, t)
	}
}

func initTest(t *testing.T) {
	var err error

	path := os.Getenv("GOPATH") + "/src/"

	for i := range testTable {
		paths := &configuration.ConfigPaths {
			ConfigFilePath: path + testTable[i].path,
			ConfigFileEmails: path + pathMails, 
		}

		if testTable[i].config, err = configuration.Init(paths); err != nil {
			t.Fatalf("La inicialización de la configuración falló; %s", err)
		} 

		if testTable[i].res, err = Init(testTable[i].config); err != nil {
			t.Fatalf("La inicialización del árbol de correos falló; %s", err)
		}
	}
}

func cmpTree(name string, node1 *MailList, node2 *MailList, t *testing.T) {
	if node1.Children != nil && node2.Children != nil {
		if len(node1.Children) != len(node2.Children) {
			t.Fatalf("El número de hijos del nodo %s no concuerdan", name)
		}

		for k1, v1 := range node1.Children {
			ok := false
			var k2 string
			var v2 *MailList
			for k2, v2 = range node2.Children {
				if k1 == k2 {
					if v1.Father != node1 || v2.Father != node2 {
						t.Fatalf("El padre del nodo %s no concuerda", k1)	
					}

					if len(v1.Mails) != len(v2.Mails) {
						t.Fatalf("Los emails de los nodos %s no concuerdan", k1)
					}

					for i := range v1.Mails {
						okMail := false
						for h := range v2.Mails {
							if v1.Mails[i] == v2.Mails[h] {
								okMail = true
								break
							}
						}

						if !okMail {
							t.Fatalf("No se encontró el email %s en el nodo %s", v1.Mails[i], k1)
						}
					}

					ok = true
					break
				}
			}

			if ok {
				cmpTree(k1, v1, v2, t)
			} else {
				t.Fatalf("En el nodo %s no se encontró el hijo %s", name, k1)	
			}
		}
	} else if !(node1.Children == nil && node2.Children == nil) {
		t.Fatalf("El número de hijos del nodo %s no concuerdan", name)
	}
}
