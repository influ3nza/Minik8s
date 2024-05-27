import base64
from PIL import Image, ImageDraw, ImageFont
from io import BytesIO


def run(image, x, y, r, g, b, mark, mysqlIp, rate, threshold1, threshold2, kernel_x, kernel_y, status):
    img = image.split(",")[1]
    img_b = base64.b64decode(img)
    image_ = Image.open(BytesIO(img_b))
    draw = ImageDraw.Draw(image_)
    try:
        font = ImageFont.truetype("arial.ttf", 40)  # 使用系统字体
    except IOError:
        font = ImageFont.load_default()

    draw.text((x, y), mark, (r, g, b), font=font)

    buffered = BytesIO()
    image_.save(buffered, format="PNG")
    image_.save('watermarked_image.png')

    img_str = base64.b64encode(buffered.getvalue())

    base64_str_with_watermark = f"data:image/png;base64,{img_str.decode()}"
    return {
        "image": base64_str_with_watermark,
        "mark": "water_mark",
        "x": x,
        "y": y,
        "r": r,
        "g": g,
        "b": b,
        "mysqlIp": mysqlIp,
        "rate": rate,
        "threshold1": threshold1,
        "threshold2": threshold2,
        "kernel_x": kernel_x,
        "kernel_y": kernel_y,
        "status": status
    }

# print(base64_str_with_watermark)

# print(img_b[0:20])
