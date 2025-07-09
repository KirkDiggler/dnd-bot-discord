# Debug Trace: What Should Happen

## 1. User selects a race from dropdown

Discord sends:
```json
{
  "type": 3,  // MESSAGE_COMPONENT
  "custom_id": "creation:preview_race:char_123",
  "values": ["elf"]
}
```

## 2. V2 Pipeline receives interaction

```go
// In shouldUseV2
customID := "creation:preview_race:char_123"
// Returns true because it starts with "creation:"
```

## 3. Pipeline executes

```go
pipeline.Execute(ctx, session, interaction)
// Calls each router's CanHandle method
```

## 4. Character Creation Router checks

```go
// Router domain is "creation"
// Parses customID: "creation:preview_race:char_123"
parsed := {
  Domain: "creation",
  Action: "preview_race", 
  Target: "char_123"
}
// Domain matches! Returns pattern: "component:preview_race"
```

## 5. Router looks up handler

```go
handlers["component:preview_race"] = HandleRacePreview
// Executes HandleRacePreview
```

## The Problem

If you're seeing "interaction failed" with no logs, it means:

1. The interaction isn't reaching the v2 pipeline (shouldUseV2 returns false)
2. OR the custom ID isn't being parsed correctly
3. OR the handler is returning an error

Let's add more debug logging to find out.