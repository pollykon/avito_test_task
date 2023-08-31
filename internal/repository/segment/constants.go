package segment

import "errors"

var ErrSegmentAlreadyExists = errors.New("segment already exists")
var ErrSegmentNotExist = errors.New("segment doesn't exists")
var ErrUserAlreadyInSegment = errors.New("user already in segment")
