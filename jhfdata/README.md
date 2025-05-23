# The Hershey Font data

## Source

The data in the [`hershey/`](hershey) sub-directory was downloaded
from

  https://emergent.unpythonic.net/software/hershey

and unpacked from the file:

  https://media.unpythonic.net/emergent-files/software/hershey/hershey.zip

It is covered by, and any update to the files in this directory will
continue to be covered by, the following license for distribution:

```
USE RESTRICTION:
	This distribution of the Hershey Fonts may be used by anyone for
	any purpose, commercial or otherwise, providing that:
		1. The following acknowledgements must be distributed with
			the font data:
			- The Hershey Fonts were originally created by Dr.
				A. V. Hershey while working at the U. S.
				National Bureau of Standards.
			- The format of the Font data in this distribution
				was originally created by
					James Hurt
					Cognition, Inc.
					900 Technology Park Drive
					Billerica, MA 01821
					(mit-eddie!ci-dandelion!hurt)
		2. The font data in this distribution may be converted into
			any other format *EXCEPT* the format distributed by
			the U.S. NTIS (which organization holds the rights
			to the distribution and use of the font data in that
			particular format). Not that anybody would really
			*want* to use their format... each point is described
			in eight bytes as "xxx yyy:", where xxx and yyy are
			the coordinate values as ASCII numbers.
```

## The `jhf` format

The format of the `*.jhf` files is described as follows:

The structure is bascially as follows: each character consists of a
number 1->4000 (not all used) in column 0:4, the number of character
pairs in columns 5:7 (pen up+down, " R", counts as a vertex), the left
hand position in column 8, the right hand position in column 9, and
finally the vertices in single character pairs. All coordinates are
given relative to the ascii value of 'R'. If the coordinate value is "
R" that indicates a pen up operation.  As an example consider the 8th
symbol

8 9MWOMOV RUMUV ROQUQ
It has 9 coordinate pairs (this includes the left and right position).
The left position is 'M' - 'R' = -5
The right position is 'W' - 'R' = 5
The first coordinate is "OM" = (-3,-5)
The second coordinate is "OV" = (-3,4)
Raise the pen " R"
Move to "UM" = (3,-5)
Draw to "UV" = (3,4)
Raise the pen " R"
Move to "OQ" = (-3,-1)
Draw to "UQ" = (3,-1)
Drawing this out on a piece of paper will reveal it represents an 'H'.

# Improvements

Note, the majority of characters in these files are assigned the
number 12345, making the fonts somewhat unusable. We provide a
mechanism to assign them to their near match utf8 code based on the
utf8/<font>.dec tables we've created for this present
distribution. Please help improve these tables via pull requests or a
bug.
