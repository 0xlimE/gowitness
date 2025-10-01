# Intro Flow Documentation

## Overview
The intro flow is a 3-slide onboarding experience that welcomes new users to the Gowitness server interface. It's shown once when users first access the application (after authentication) and can be skipped at any time.

## Files Created/Modified

### New Files:
1. **`/web/ui/src/pages/intro/Intro.tsx`** - Main intro component with 3 slides
2. **`/web/ui/src/lib/cookies.ts`** - Cookie utilities for managing the "has_seen_intro" flag

### Modified Files:
1. **`/web/ui/src/main.tsx`** - Updated router to include intro route and check cookie on dashboard load

## How It Works

1. When a user first accesses the dashboard (`/`), a loader function checks for the `has_seen_intro` cookie
2. If the cookie is not set, the user is redirected to `/intro`
3. The intro page displays 3 slides:
   - **Slide 1**: Welcome message and overview
   - **Slide 2**: Key features
   - **Slide 3**: Getting started guide
4. Users can:
   - Navigate forward/backward through slides
   - Skip the intro entirely
   - Complete the intro (on the last slide)
5. When completed or skipped, the `has_seen_intro` cookie is set and the user is redirected to the dashboard

## Cookie Details
- **Name**: `has_seen_intro`
- **Value**: `true`
- **Max Age**: 31536000 seconds (1 year)
- **Path**: `/` (available across the entire application)

## Customization

### Updating Slide Content
Edit the `introSlides` array in `/web/ui/src/pages/intro/Intro.tsx`:

```tsx
const introSlides = [
  {
    title: 'Your Title',
    description: 'Your description',
    content: (
      <div>Your JSX content here</div>
    ),
  },
  // Add more slides...
];
```

### Resetting the Intro for Testing
To see the intro again during development:
1. Open browser DevTools
2. Go to Application/Storage > Cookies
3. Delete the `has_seen_intro` cookie
4. Refresh the page

Or programmatically in the browser console:
```javascript
document.cookie = 'has_seen_intro=; path=/; max-age=0';
```

## Future Enhancements

Possible improvements for the intro flow:
- Add animations between slides (e.g., fade, slide transitions)
- Make slide count dynamic based on array length
- Add keyboard navigation (arrow keys)
- Include interactive elements or demos
- Add analytics to track which slides users skip/complete
- Support for multiple intro versions (versioned cookies)
