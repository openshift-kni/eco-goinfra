package msg

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUndefinedCrdObjectErrString(t *testing.T) {
	assert.Equal(t, "can not redefine the undefined crName", UndefinedCrdObjectErrString("crName"))
}

func TestFailToUpdateNotification(t *testing.T) {
	//nolint:lll
	assert.Equal(t, "Failed to update the crName object objName in namespace nsName. Note: Force flag set, executed delete/create methods instead", FailToUpdateNotification("crName", "objName", "nsName"))
}

func TestFailToUpdateError(t *testing.T) {
	//nolint:lll
	assert.Equal(t, "Failed to update the crName object objName in namespace nsName, due to error in delete function", FailToUpdateError("crName", "objName", "nsName"))
}
