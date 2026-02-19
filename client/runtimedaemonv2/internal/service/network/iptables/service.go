package iptables

import (
	"context"
	"fmt"
	"strings"
)

const (
	appendRuleFlag = "-A"
	changePolicy   = "-P"
)

const minRulePartsCount = 3

func (s service) Set(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, table := range s.config.Tables {
		// create new user chains
		for _, chain := range table.UserChains {
			if err := s.iptablesAPI.NewChain(table.Name, chain); err != nil {
				log.Warningf("failed to create new chain (%v): %v", chain, err)
			}
		}

		// appends rules & changes default policies
		for _, rule := range table.Rules {
			split, err := splitRule(rule)
			if err != nil {
				return err
			}

			// iptables rule format [flag] [chain] [target]
			// get chain from rule
			chain := split[1]

			switch {
			case strings.HasPrefix(rule, appendRuleFlag):
				rulespec := split[2:]

				if err := s.iptablesAPI.AppendUnique(table.Name, chain, rulespec...); err != nil {
					return fmt.Errorf("failed to create new rule (%v): %v", rule, err)
				}

			case strings.HasPrefix(rule, changePolicy):
				target := strings.Join(split[2:], " ")

				if err := s.iptablesAPI.ChangePolicy(table.Name, chain, target); err != nil {
					return fmt.Errorf("failed to change policy (%v): %v", rule, err)
				}

			default:
				log.Errorf("unknown rule: %v", rule)
			}

		}
	}

	return nil
}

func (s service) Discard(ctx context.Context) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for _, table := range s.config.Tables {
		// Discard rules
		for _, rule := range table.Rules {
			if strings.HasPrefix(rule, appendRuleFlag) {
				split, err := splitRule(rule)
				if err != nil {
					return err
				}

				if err := s.iptablesAPI.DeleteIfExists(table.Name, split[1], split[2:]...); err != nil {
					return fmt.Errorf("failed to delete iptables rule: %s", rule)
				}
			}
		}

		// clear and delete chains
		for _, chain := range table.UserChains {
			if err := s.iptablesAPI.ClearAndDeleteChain(table.Name, chain); err != nil {
				return fmt.Errorf("failed to clear iptables chain: %s", chain)
			}
		}
	}

	return nil
}

func (s service) IsCorrect(ctx context.Context, mustBeSet bool) (bool, error) {
	log.Info("[iptables is correct]: called")

	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Info("[iptables is correct]: started")

	for _, table := range s.config.Tables {

		// check if user chains exists
		chains, err := s.iptablesAPI.ListChains(table.Name)
		if err != nil {
			return false, fmt.Errorf("failed to list iptables chains (%v): %w", table, err)
		}

		for _, chain := range table.UserChains {
			if contains(chains, chain) != mustBeSet {
				return false, nil
			}
		}

		// check if rules exists
		rulesExist := make([]bool, len(table.Rules))

		for _, chain := range chains {

			rules, err := s.iptablesAPI.List(table.Name, chain)
			if err != nil {
				return false, fmt.Errorf("failed to list iptables chains (%v): %w", chain, err)
			}

			for _, rule := range rules {
				ind, contains := containsIn(table.Rules, rule)

				if !contains {
					continue
				}

				rulesExist[ind] = true
			}

		}

		for i, exist := range rulesExist {
			if exist != mustBeSet {
				// skip default policy check if mustBeSet == false, because default policy should not be discarded
				if !mustBeSet && strings.HasPrefix(table.Rules[i], changePolicy) {
					continue
				}

				return false, nil
			}
		}

	}

	return true, nil
}

func contains(slice []string, s string) bool {
	for _, e := range slice {
		if s == e {
			return true
		}
	}

	return false
}

func containsIn(slice []string, s string) (int, bool) {
	for i, e := range slice {
		if s == e {
			return i, true
		}
	}

	return -1, false
}

func splitRule(rule string) ([]string, error) {
	split := strings.Split(rule, " ")

	if len(split) < minRulePartsCount {
		return nil, fmt.Errorf("failed to parse iptables rule: %s", rule)
	}

	return split, nil
}
