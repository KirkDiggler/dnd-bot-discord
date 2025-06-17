package character

// This file demonstrates how the Discord handlers would integrate with the service layer
// It's not meant to be compiled - just shows the pattern

/*
Example of how a Discord handler would use the character service:

type CharacterCreationState struct {
    UserID        string
    GuildID       string
    RaceKey       string
    ClassKey      string
    AbilityScores map[string]int
    Proficiencies []string
    Equipment     []string
}

// In the Discord handler when user submits character name:
func handleCharacterNameSubmit(s *discordgo.Session, i *discordgo.InteractionCreate, state *CharacterCreationState) error {
    // Extract character name from modal
    characterName := i.ModalSubmitData().Components[0].(*discordgo.ActionsRow).Components[0].(*discordgo.TextInput).Value

    // Create the character using the service
    input := &character.CreateCharacterInput{
        UserID:        state.UserID,
        RealmID:       state.GuildID,  // Discord guild -> domain realm
        Name:          characterName,
        RaceKey:       state.RaceKey,
        ClassKey:      state.ClassKey,
        AbilityScores: state.AbilityScores,
        Proficiencies: state.Proficiencies,
        Equipment:     state.Equipment,
    }

    output, err := characterService.CreateCharacter(context.Background(), input)
    if err != nil {
        // Handle error - show to user
        return respondWithError(s, i, "Failed to create character: " + err.Error())
    }

    // Show success message
    embed := &discordgo.MessageEmbed{
        Title:       "Character Created!",
        Description: fmt.Sprintf("Your character **%s** has been created successfully!", output.Character.Name),
        Color:       0x00FF00,
        Fields: []*discordgo.MessageEmbedField{
            {
                Name:  "Character ID",
                Value: output.Character.ID,
                Inline: true,
            },
            {
                Name:  "Race",
                Value: state.RaceKey,
                Inline: true,
            },
            {
                Name:  "Class",
                Value: state.ClassKey,
                Inline: true,
            },
        },
    }

    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseUpdateMessage,
        Data: &discordgo.InteractionResponseData{
            Embeds: []*discordgo.MessageEmbed{embed},
        },
    })
}

// When showing proficiency choices:
func handleShowProficiencyChoices(s *discordgo.Session, i *discordgo.InteractionCreate, raceKey, classKey string) error {
    // Get available choices from service
    input := &character.ResolveChoicesInput{
        RaceKey:  raceKey,
        ClassKey: classKey,
    }

    output, err := characterService.ResolveChoices(context.Background(), input)
    if err != nil {
        return respondWithError(s, i, "Failed to get choices: " + err.Error())
    }

    // Convert service output to Discord components
    var components []discordgo.MessageComponent

    for _, choice := range output.ProficiencyChoices {
        if len(choice.Options) <= 25 { // Discord limit
            // Create dropdown
            options := make([]discordgo.SelectMenuOption, len(choice.Options))
            for j, opt := range choice.Options {
                options[j] = discordgo.SelectMenuOption{
                    Label: opt.Name,
                    Value: opt.Key,
                }
            }

            components = append(components, discordgo.ActionsRow{
                Components: []discordgo.MessageComponent{
                    discordgo.SelectMenu{
                        CustomID:    fmt.Sprintf("prof_choice:%s", choice.ID),
                        Placeholder: choice.Name,
                        MinValues:   &choice.Choose,
                        MaxValues:   choice.Choose,
                        Options:     options,
                    },
                },
            })
        } else {
            // Too many options - use buttons to show sub-menus
            // Implementation details...
        }
    }

    // Show the choices to the user
    return s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
        Type: discordgo.InteractionResponseUpdateMessage,
        Data: &discordgo.InteractionResponseData{
            Content:    "Select your proficiencies:",
            Components: components,
        },
    })
}
*/
