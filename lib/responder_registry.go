package lib

const ActionsResponder = "actions"
const ModalResponder = "modal"

func NewResponderRegistry() *ResponderRegistry {
	return &ResponderRegistry{
		responders: map[string]Responder{},
	}
}

type ResponderRegistry struct {
	responders map[string]Responder
}

type Responder struct {
	Name        string
	Respond     InteractionAction
	Authorizers []Authorizer
}

func (sr *ResponderRegistry) Register(r Responder) {
	if _, ok := sr.responders[r.Name]; ok {
		panic("Responder already registered: " + r.Name)
	}
	sr.responders[r.Name] = r
}

func (sr *ResponderRegistry) GetAll() map[string]Responder {
	return sr.responders
}
