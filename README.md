# picture-to-array

将图片转换为c数组

## 用法

以下划线`_`或者小数点`.`开头的文件将会被忽略。

### 单文件

`app.exe -in abc/assets -out out/carrays`

`abc/assets/x/y/z/pic.png` 导出名称将会是 `x_y_z_pic_bmp`

### 数组

当目录名称中含有`[array]`时, 目录中的图片将会以数组形式导出。

```txt
abc/assets/ddd/example[array]/0.png
abc/assets/ddd/example[array]/1.png
abc/assets/ddd/example[array]/a.png
abc/assets/ddd/example[array]/b.png
```

以上文件将会导出至 `ddd_example_array[]` 数组。

## bitmap.h

```c
#ifndef BITMAP_H
#define BITMAP_H

#include <stdint.h>

typedef struct SPOS {
	int16_t x;
	int16_t y;
} sPOS;

typedef struct SBITMAP {
	uint8_t w;
	uint8_t h;
	const uint8_t* map;
} sBITMAP;

typedef struct SRECT {
	int16_t x;
	int16_t y;
	int16_t w;
	int16_t h;
} sRECT;

typedef enum BLENDMODE {
	REPLACE,
	OR,
	ERASE,
	AND,
	NOT,
	XOR,
	XNOR,
} eBlendMode;
#endif
```
