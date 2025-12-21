package cron

import (
	"github.com/robfig/cron/v3"
)

type Parser struct {
	parser cron.Parser
}

func NewParser() *Parser {
	return &Parser{
		parser: cron.NewParser(
			cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
		),
	}
}

func (p *Parser) Parse(expression string) (cron.Schedule, error) {
	return p.parser.Parse(expression)
}

func (p *Parser) Validate(expression string) error {
	_, err := p.parser.Parse(expression)
	return err
}
