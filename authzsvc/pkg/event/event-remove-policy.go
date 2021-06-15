package event

const EventRemovePolicy EventName = "EventRemovePolicy"

// register the event to the registry
func init() {
	Registry.register(EventRemovePolicy, EventInfo{
		ReqChan: "authzsvc.EventRemovePolicy",
		isValidPayload: func(i interface{}) bool {
			_, ok := i.(EventRemovePolicyPayload)
			return ok
		},
	})
}

type EventRemovePolicyPayload struct {
	*EventUpsertPolicyPayload
}
