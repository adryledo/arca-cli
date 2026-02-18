def process(items):
    result = []
    for i in items:
        if i > 10:
            result.append(i * 2)
        else:
            result.append(i)
    return result