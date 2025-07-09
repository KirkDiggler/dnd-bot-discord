package middleware

import (
	"github.com/KirkDiggler/dnd-bot-discord/internal/discord/v2/core"
	"github.com/bwmarrin/discordgo"
)

// AuthConfig configures authorization behavior
type AuthConfig struct {
	// RequireGuildMember requires user to be a guild member
	RequireGuildMember bool

	// RequiredRoles lists role IDs user must have (any of)
	RequiredRoles []string

	// RequiredPermissions lists permissions user must have
	RequiredPermissions int64

	// UserWhitelist allows specific users regardless of other checks
	UserWhitelist []string

	// UserBlacklist blocks specific users
	UserBlacklist []string

	// CustomChecker allows custom authorization logic
	CustomChecker AuthChecker
}

// AuthChecker is a custom authorization function
type AuthChecker func(ctx *core.InteractionContext) (bool, string)

// AuthorizationMiddleware checks if user is authorized
func AuthorizationMiddleware(config *AuthConfig) core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			// Check blacklist first
			if isBlacklisted(ctx.UserID, config.UserBlacklist) {
				return unauthorizedResponse("You are not authorized to use this command."), nil
			}

			// Check whitelist - bypass other checks
			if isWhitelisted(ctx.UserID, config.UserWhitelist) {
				return next.Handle(ctx)
			}

			// Check guild member requirement
			if config.RequireGuildMember && ctx.GuildID == "" {
				return unauthorizedResponse("This command can only be used in a server."), nil
			}

			// Check roles
			if len(config.RequiredRoles) > 0 && !hasRequiredRole(ctx, config.RequiredRoles) {
				return unauthorizedResponse("You don't have the required role to use this command."), nil
			}

			// Check permissions
			if config.RequiredPermissions > 0 && !hasRequiredPermissions(ctx, config.RequiredPermissions) {
				return unauthorizedResponse("You don't have the required permissions to use this command."), nil
			}

			// Run custom checker
			if config.CustomChecker != nil {
				if allowed, reason := config.CustomChecker(ctx); !allowed {
					return unauthorizedResponse(reason), nil
				}
			}

			// User is authorized
			return next.Handle(ctx)
		})
	}
}

// RoleRequiredMiddleware requires user to have specific role
func RoleRequiredMiddleware(roleIDs ...string) core.Middleware {
	return AuthorizationMiddleware(&AuthConfig{
		RequiredRoles: roleIDs,
	})
}

// PermissionRequiredMiddleware requires user to have specific permissions
func PermissionRequiredMiddleware(permissions int64) core.Middleware {
	return AuthorizationMiddleware(&AuthConfig{
		RequiredPermissions: permissions,
	})
}

// OwnerOnlyMiddleware restricts to bot owner or server owner
func OwnerOnlyMiddleware(botOwnerID string) core.Middleware {
	return AuthorizationMiddleware(&AuthConfig{
		CustomChecker: func(ctx *core.InteractionContext) (bool, string) {
			// Check if bot owner
			if ctx.UserID == botOwnerID {
				return true, ""
			}

			// Check if server owner
			if ctx.GuildID != "" && ctx.Interaction.Member != nil {
				guild, err := ctx.Session.Guild(ctx.GuildID)
				if err == nil && guild.OwnerID == ctx.UserID {
					return true, ""
				}
			}

			return false, "This command is restricted to bot/server owners."
		},
	})
}

// DomainAuthMiddleware applies auth rules based on command domain
func DomainAuthMiddleware(rules map[string]*AuthConfig) core.Middleware {
	return func(next core.Handler) core.Handler {
		return core.HandlerFunc(func(ctx *core.InteractionContext) (*core.HandlerResult, error) {
			var domain string

			// Extract domain from interaction
			if ctx.IsCommand() {
				domain = ctx.GetCommandName()
			} else if ctx.IsComponent() {
				if parsed, err := core.ParseCustomID(ctx.GetCustomID()); err == nil {
					domain = parsed.Domain
				}
			}

			// Check if we have rules for this domain
			if config, ok := rules[domain]; ok {
				// Apply domain-specific auth
				authHandler := AuthorizationMiddleware(config)(next)
				return authHandler.Handle(ctx)
			}

			// No specific rules, allow through
			return next.Handle(ctx)
		})
	}
}

// isBlacklisted checks if user is in blacklist
func isBlacklisted(userID string, blacklist []string) bool {
	for _, id := range blacklist {
		if id == userID {
			return true
		}
	}
	return false
}

// isWhitelisted checks if user is in whitelist
func isWhitelisted(userID string, whitelist []string) bool {
	for _, id := range whitelist {
		if id == userID {
			return true
		}
	}
	return false
}

// hasRequiredRole checks if member has any of the required roles
func hasRequiredRole(ctx *core.InteractionContext, requiredRoles []string) bool {
	if ctx.Member == nil {
		return false
	}

	for _, memberRole := range ctx.Member.Roles {
		for _, requiredRole := range requiredRoles {
			if memberRole == requiredRole {
				return true
			}
		}
	}

	return false
}

// hasRequiredPermissions checks if member has required permissions
func hasRequiredPermissions(ctx *core.InteractionContext, required int64) bool {
	if ctx.Member == nil || ctx.GuildID == "" {
		return false
	}

	// Get member permissions
	permissions, err := ctx.Session.UserChannelPermissions(ctx.UserID, ctx.ChannelID)
	if err != nil {
		return false
	}

	// Check if has admin (bypasses all permission checks)
	if permissions&discordgo.PermissionAdministrator != 0 {
		return true
	}

	// Check specific permissions
	return permissions&required == required
}

// unauthorizedResponse creates an unauthorized error response
func unauthorizedResponse(message string) *core.HandlerResult {
	return &core.HandlerResult{
		Response: core.NewEphemeralResponse("❌ " + message),
	}
}
