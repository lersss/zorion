package batch

import (
	"testing"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestResolve_SimpleSum(t *testing.T) {
	locID := uuid.New()
	effects := []Effect{
		{ID: uuid.New(), ResourceType: "wood", Delta: 10},
		{ID: uuid.New(), ResourceType: "wood", Delta: -5},
		{ID: uuid.New(), ResourceType: "wood", Delta: -3},
	}
	state := ResourceState{"wood": 100}
	config := DefaultResolverConfig()
	
	bundle, err := Resolve(locID, 1, state, effects, config)
	assert.NoError(t, err)
	
	// Итоговая дельта: 10 - 5 - 3 = +2
	assert.Equal(t, int64(2), bundle.Effects[0].Delta)
	assert.Equal(t, 1, len(bundle.Effects))
}

func TestResolve_NotEnoughResource(t *testing.T) {
	locID := uuid.New()
	effects := []Effect{
		{ID: uuid.New(), ResourceType: "wood", Delta: -150},
	}
	state := ResourceState{"wood": 100}
	config := DefaultResolverConfig()
	
	bundle, err := Resolve(locID, 1, state, effects, config)
	assert.NoError(t, err)
	
	// Должен быть урезан до -100 (не уходим в минус)
	assert.Equal(t, int64(-100), bundle.Effects[0].Delta)
}