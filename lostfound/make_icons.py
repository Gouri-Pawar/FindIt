from PIL import Image, ImageDraw
import os


def make_icon(size, path):
    # blue to purple gradient background
    img = Image.new("RGB", (size, size))
    px = img.load()
    for y in range(size):
        for x in range(size):
            t = (x + y) / (2 * size)
            px[x, y] = (int(37 + 87 * t), int(99 - 41 * t), int(235 + 2 * t))

    # white magnifying glass with a tick inside
    d = ImageDraw.Draw(img)
    s = size
    cx, cy, r = s * 0.44, s * 0.44, s * 0.17
    w = int(s * 0.055)
    d.ellipse([cx - r, cy - r, cx + r, cy + r], outline="white", width=w)
    d.line([cx + r * 0.72, cy + r * 0.72, s * 0.70, s * 0.70], fill="white", width=w)
    d.line(
        [cx - r * 0.45, cy, cx - r * 0.1, cy + r * 0.4, cx + r * 0.5, cy - r * 0.4],
        fill="white", width=int(w * 0.7), joint="curve",
    )
    img.save(path)


os.makedirs("web/icons", exist_ok=True)
make_icon(192, "web/icons/icon-192.png")
make_icon(512, "web/icons/icon-512.png")
print("Icons created in web/icons")