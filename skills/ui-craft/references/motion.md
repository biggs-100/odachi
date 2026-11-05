# Motion

First the question: feedback, continuity, hierarchy or story; if it explains none,
it does not go. 100ms feedback, 150-200 exit, 200-300 enter, 300-500 for a large
transition; ease-out to enter, ease-in to exit, ease-in-out to move, never linear
except progress and loops. Nothing over 500ms without a reason a person would say
out loud. Gestures get springs with 1:1 tracking and velocity handoff, and every
animation is interruptible. One lead element, the rest staggered 20-50ms, at most
two properties; layout uses FLIP and pages use View Transitions. Transform and
opacity only; respect prefers-reduced-motion, and no parallax, scroll-jacking or
autoplay carousel.

## The reduced-motion pass

Reduced motion removes travel, not meaning: a panel can appear instead of sliding,
and a state change still needs its cue. Test the screen with the setting on before
you call the motion done.

## The two-property rule

Animate transform and opacity. A property that triggers layout or paint turns a
cheap transition into a frame budget the page cannot pay, and the fix is a
different composition, not a shorter duration.
