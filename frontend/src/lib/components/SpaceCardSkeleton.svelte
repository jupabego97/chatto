<!--
@component

Skeleton placeholder for a SpaceCard, shown while the Browse Spaces page is loading.
Mirrors the SpaceCard layout with animated shimmer placeholders.
-->

<article class="flex flex-col overflow-hidden rounded-xl border border-border bg-background">
  <!-- Banner skeleton -->
  <div class="relative shrink-0">
    <div class="skeleton aspect-4/3 w-full"></div>

    <!-- Logo skeleton -->
    <div
      class="skeleton absolute -bottom-6 left-4 z-10 h-12 w-12 rounded-xl border-2 border-background shadow-md"
    ></div>
  </div>

  <!-- Content skeleton -->
  <div class="flex flex-1 flex-col p-4 pt-8">
    <!-- Title -->
    <div class="skeleton h-5 w-2/3 rounded"></div>

    <!-- Description lines -->
    <div class="mt-2 flex flex-col gap-1.5">
      <div class="skeleton h-3.5 w-full rounded"></div>
      <div class="skeleton h-3.5 w-4/5 rounded"></div>
    </div>

    <!-- Footer -->
    <div class="mt-auto flex items-center justify-between pt-4">
      <div class="skeleton h-4 w-24 rounded"></div>
      <div class="skeleton h-8 w-16 rounded-md"></div>
    </div>
  </div>
</article>
