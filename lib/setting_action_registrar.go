package lib

func NewSettingActionRegistrar() *SettingActionRegistrar {
	return &SettingActionRegistrar{
		responders: map[string]SettingAction{},
	}
}

type SettingActionRegistrar struct {
	responders map[string]SettingAction
}

type SettingAction struct {
	Name        string
	Respond     InteractionAction
	Authorizers []Authorizer
}

func (sar *SettingActionRegistrar) Register(sa SettingAction) {
	if _, ok := sar.responders[sa.Name]; ok {
		panic("Setting action already registered: " + sa.Name)
	}
	sar.responders[sa.Name] = sa
}

func (sar *SettingActionRegistrar) GetAll() map[string]SettingAction {
	return sar.responders
}
