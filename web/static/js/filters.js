// web/static/js/filters.js

export const filterState = {
    hasPlanets: false,
    hasLife: false,
    hasHabitable: false,
    planetType: '',
    resourceCategory: '',
};

export function resetFilters() {
    filterState.hasPlanets = false;
    filterState.hasLife = false;
    filterState.hasHabitable = false;
    filterState.planetType = '';
    filterState.resourceCategory = '';
}

export function applyFiltersFromUI() {
    filterState.hasPlanets = document.getElementById('filter-has-planets').checked;
    filterState.hasLife = document.getElementById('filter-life').checked;
    filterState.hasHabitable = document.getElementById('filter-habitable').checked;
    filterState.planetType = document.getElementById('filter-planet-type').value;
    filterState.resourceCategory = document.getElementById('filter-resource').value;
}