# Brand

Source of truth for the Chiro mark. Colors: tile `#5B7CF6`, glyph white.

| File | Use |
| --- | --- |
| `logo.svg` | Primary logo (tile + C). Scales to any size. |
| `logo-mark.svg` | C glyph only, transparent background, for placing on your own surface. |
| `favicon.svg`, `favicon.ico` | Browser tab icon (SVG first, ICO fallback). |
| `icon-1024.png` | App store / master raster. |
| `icon-512.png`, `icon-192.png` | PWA manifest icons. |
| `icon-180.png` | iOS home screen (`apple-touch-icon`). |
| `icon-32.png` | Small raster fallback. |
| `icon-maskable.svg`, `icon-maskable-512.png` | Android adaptive/maskable icon (full bleed, glyph inside the safe zone). |

PNGs are generated from the SVGs:

```sh
rsvg-convert -w <size> -h <size> logo.svg -o icon-<size>.png
```
