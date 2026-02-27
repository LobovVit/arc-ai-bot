package llm

type Stub struct{}

func (s *Stub) Chat(system string, user string) (string, error) {
	return "Я не знаю.", nil
}
