// TEMPORARY FIX: Log date filters to console
console.log('Sequence Device Report - Date Filters:', {
    startDate: sequenceStartDate,
    endDate: sequenceEndDate,
    url: url
});

// Additional debug to understand the issue
console.log('Current date filters when opening device report:', {
    sequenceStartDate: sequenceStartDate || 'Not set',
    sequenceEndDate: sequenceEndDate || 'Not set',
    defaulting_to_today: !sequenceStartDate && !sequenceEndDate
});
