
function UpdateDateDiffAsText($scope) {
  var today = new Date();
  var diff = Math.floor(today.getTime() - $scope.dateValue.getTime());
  var day = 1000 * 60 * 60 * 24;

  var days = Math.floor(diff/day);

  var dateDiffFromTodayAsText = ""
    if (days == 0) {
      dateDiffFromTodayAsText = "Today";
    }
    else if (days ==1) {
      dateDiffFromTodayAsText = "1 day old";
    } else {
      dateDiffFromTodayAsText = days + " days old";
    }
  $scope.dateDiffFromTodayAsText = dateDiffFromTodayAsText;
}
