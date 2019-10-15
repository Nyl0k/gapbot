package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func getRole(s *discordgo.Session, u *discordgo.User, m *discordgo.Message, roleName string) bool {
	// see what the user permission level is
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting roles: %s", err)
		return false
	}
	mem, err := s.GuildMember(m.GuildID, u.ID)
	if err != nil {
		fmt.Printf("Error getting members: %s", err)
		return false
	}
	for _, role := range roles {
		if role.Name == roleName {
			// check and see if user has role
			for _, v := range mem.Roles {
				if role.ID == v {
					return true
				}
			}
		}
	}
	return false
}

func getAdmin(s *discordgo.Session, m *discordgo.Message) bool {
	return getRole(s, m.Author, m, configData.ModRoleName)
}

func Role(s *discordgo.Session, m *discordgo.MessageCreate, removing bool) {
	// if an admin is modifying another persons' role
	if getAdmin(s, m.Message) && len(m.Mentions) > 0 {
		split := strings.SplitN(strings.TrimPrefix(m.Content, configData.Prefix), " ", 3)
		if len(split) != 3 {
			s.ChannelMessageSend(m.ChannelID, "Make sure to specify a role after the mention.")
			return
		}
		role := split[2]
		AddDelRole(m.Message, m.Mentions[0], s, role, removing)
		return
	}
	split := strings.SplitN(strings.TrimPrefix(m.Content, configData.Prefix), " ", 2)
	if len(split) < 2 {
		s.ChannelMessageSend(m.ChannelID, "Make sure to specify a role.")
		return
	}
	role := split[1]
	AddDelRole(m.Message, m.Author, s, role, removing)
}

func AddDelRole(m *discordgo.Message, u *discordgo.User, s *discordgo.Session, roleName string, removing bool) {
	// first see if the user has permissions to change the role
	availableRoles, err := getAvailableRoles(s, m, m.Author)
	if err != nil {
		fmt.Printf("Error getting available roles: %s", err)
		return
	}
	found := false
	for _, role := range availableRoles {
		if roleName == strings.ToLower(role.Name) {
			found = true
		}
	}
	if !found {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("You don't have permission to change role ``%s``.", roleName))
		return
	}
	g, err := s.Guild(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting guild: %s", err)
		return
	}
	rs, err := s.GuildRoles(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting roles: %s", err)
		return
	}
	for _, role := range rs {
		if strings.ToLower(role.Name) == roleName {
			if removing {
				err := s.GuildMemberRoleRemove(g.ID, u.ID, role.ID)
				if err != nil {
					fmt.Printf("Error removing role %s to %s", roleName, m.Author.Username)
					return
				}
				s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role ``%s`` successfully removed.", roleName))
				return
			}
			err := s.GuildMemberRoleAdd(g.ID, u.ID, role.ID)
			if err != nil {
				fmt.Printf("Error adding role %s to %s", roleName, m.Author.Username)
				return
			}
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role ``%s`` successfully added.", roleName))
			return
		}
	}
	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Role ``%s`` not found. Please try again or contact an admin.", roleName))
}

func getGuildRoles(s *discordgo.Session, m *discordgo.Message) ([]*discordgo.Role, error) {
	roles, err := s.GuildRoles(m.GuildID)
	if err != nil {
		return nil, err
	}
	sort.SliceStable(roles, func(i, j int) bool {
		return roles[i].Position > roles[j].Position
	})
	return roles, nil
}

func getAvailableRoles(s *discordgo.Session, m *discordgo.Message, u *discordgo.User) ([]*discordgo.Role, error) {
	admin := getAdmin(s, m)
	roles := []*discordgo.Role{}
	guildRoles, err := getGuildRoles(s, m)
	if err != nil {
		return nil, err
	}
	for _, role := range guildRoles {
		if itemInSlice(role.Name, configData.EnabledRoles.UserRoles) {
			roles = append(roles, role)
		}
		if admin && itemInSlice(role.Name, configData.EnabledRoles.AdminRoles) {
			roles = append(roles, role)
		}
	}
	return roles, nil
}

func AutoAddMember(m *discordgo.Message, s *discordgo.Session, roleName string) {
	fmt.Printf("Here is the guild id: %s", m.GuildID)
	rs, err := s.GuildRoles(m.GuildID)
	if err != nil {
		fmt.Printf("Error getting roles in auto add member: %s", err)
		return
	}
	for _, role := range rs {
		if strings.ToLower(role.Name) == roleName {
			err := s.GuildMemberRoleAdd(m.GuildID, m.Author.ID, role.ID)
			if err != nil {
				fmt.Printf("Error adding role %s to %s", roleName, m.Author.Username)
				return
			}
		}
	}
}
