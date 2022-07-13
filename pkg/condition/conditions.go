package condition

import (
	"github.com/acorn-io/baaah/pkg/router"
	v1 "github.com/ekristen/alertmanager-controller/pkg/apis/alertmanager.ekristen.dev/v1"
	kclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Conditions interface {
	kclient.Object
	Conditions() *[]v1.Condition
}

func Setter(cond Conditions, resp router.Response, name string) *Callback {
	return &Callback{
		cond: cond,
		name: name,
		resp: resp,
	}
}

type Callback struct {
	name string
	cond Conditions
	resp router.Response
}

func (c *Callback) Success() {
	c.Set(v1.Condition{
		Success: true,
	})
}

func (c *Callback) Unknown(msg string) {
	c.Set(v1.Condition{
		Message:       msg,
		Transitioning: true,
	})
}

func (c *Callback) Error(err error) {
	if err == nil {
		c.Success()
		return
	}
	c.Set(v1.Condition{
		Error:   true,
		Message: err.Error(),
	})
}

func (c *Callback) Set(cond v1.Condition) {
	for i, existing := range *c.cond.Conditions() {
		if existing.Type == c.name {
			(*c.cond.Conditions())[i] = existing.Set(cond, c.cond.GetGeneration())
			return
		}
	}
	*c.cond.Conditions() = append(*c.cond.Conditions(), cond.Init(c.name, c.cond.GetGeneration()))
	c.resp.Objects(c.cond)
}
