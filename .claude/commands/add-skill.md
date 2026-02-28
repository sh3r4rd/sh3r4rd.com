Add a new skill entry to the `skills` array in `src/components/sections/SkillsSection.jsx`.

The skill to add: $ARGUMENTS

Follow the exact pattern used by existing entries:

```jsx
{
  name: "Skill Name",
  description: (<>Description text with <span className="font-semibold dark:text-white">Skill Name</span> highlighted on first mention. Additional detail about experience and usage.</>)
}
```

Rules:
- Bold the skill name on first mention using `<span className="font-semibold dark:text-white">Skill Name</span>`
- Bold other technology names mentioned using the same span pattern
- Write 2-4 sentences in a first-person professional tone matching existing entries
- Use JSX fragment syntax (`<>...</>`) for the description value
- Add the new entry at the end of the `skills` array, before the closing `]`
- Do not modify any existing entries
