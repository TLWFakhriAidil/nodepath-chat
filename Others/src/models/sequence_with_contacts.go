package models

// SequenceWithContacts represents a sequence with contacts ready for next message
type SequenceWithContacts struct {
	Sequence    Sequence
	DueContacts []ContactWithStep
}

// ContactWithStep represents a contact with their next step
type ContactWithStep struct {
	Contact  SequenceContact
	NextStep SequenceStep
}