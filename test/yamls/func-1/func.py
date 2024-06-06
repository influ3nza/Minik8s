def run(x, y):
    z = x + y
    x = x - y
    y = y - x
    print(z)
    result = {
        "x": x,
        "y": y,
        "z": z + 10
    }
    return result