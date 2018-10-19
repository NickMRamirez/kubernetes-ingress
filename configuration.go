package main

import (
	clientnative "github.com/haproxytech/client-native"
	"k8s.io/apimachinery/pkg/watch"
)

//Configuration represents k8s state
type Configuration struct {
	Namespace map[string]*Namespace
	ConfigMap *ConfigMap
	NativeAPI *clientnative.HAProxyClient
}

//Init itialize configuration
func (c *Configuration) Init(api *clientnative.HAProxyClient) {
	c.Namespace = make(map[string]*Namespace)
	c.NativeAPI = api
}

//GetNamespace returns Namespace. Creates one if not existing
func (c *Configuration) GetNamespace(name string) *Namespace {
	namespace, ok := c.Namespace[name]
	if ok {
		return namespace
	}
	newNamespace := c.NewNamespace(name)
	c.Namespace[name] = newNamespace
	return newNamespace
}

//NewNamespace returns new initialized Namespace
func (c *Configuration) NewNamespace(name string) *Namespace {
	namespace := &Namespace{
		Name:     name,
		Relevant: name == "default",
		//Annotations
		Pods:      make(map[string]*Pod),
		PodNames:  make(map[string]bool),
		Services:  make(map[string]*Service),
		Ingresses: make(map[string]*Ingress),
		Secret:    make(map[string]*Secret),
		Status:    watch.Added,
	}
	return namespace
}

//Clean cleans all the statuses of various data that was changed
//deletes them completely or just resets them if needed
func (c *Configuration) Clean() {
	for _, namespace := range c.Namespace {
		for _, data := range namespace.Ingresses {
			for _, rule := range data.Rules {
				switch rule.Status {
				case watch.Deleted:
					delete(data.Rules, rule.Host)
					continue
				default:
					rule.Status = ""
					for _, path := range rule.Paths {
						switch path.Status {
						case watch.Deleted:
							delete(rule.Paths, path.Path)
						default:
							path.Status = ""
						}
					}
				}
			}
			data.Annotations.SetStatusState("")
			switch data.Status {
			case watch.Deleted:
				delete(namespace.Ingresses, data.Name)
			default:
				data.Status = ""
			}
		}
		for _, data := range namespace.Services {
			data.Annotations.SetStatusState("")
			switch data.Status {
			case watch.Deleted:
				delete(namespace.Services, data.Name)
			default:
				data.Status = ""
			}
		}
		for _, data := range namespace.Pods {
			switch data.Status {
			case watch.Deleted:
				delete(namespace.PodNames, data.HAProxyName)
				delete(namespace.Pods, data.Name)
			default:
				data.Status = ""
			}
		}
		for _, data := range namespace.Secret {
			switch data.Status {
			case watch.Deleted:
				delete(namespace.Secret, data.Name)
			default:
				data.Status = ""
			}
		}
	}
	c.ConfigMap.Annotations.SetStatusState("")
	switch c.ConfigMap.Status {
	case watch.Deleted:
		c.ConfigMap = nil
	default:
		c.ConfigMap.Status = ""
	}
}
