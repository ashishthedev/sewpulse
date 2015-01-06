var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyAssemblyController', ['$scope', '$http', function($scope, $http) {

  function UpdateDateDiffAsText() {
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

  function UpdateTotalQty() {
    var t = 0;
    for (var i=0; i < $scope.items.length; i++) {
      t += parseInt($scope.items[i].quantity);
      console.log(t);
    }
    $scope.totalQuantityProduced = t;
  }

  $scope.removeEntry = function(index) {
    $scope.items.splice(index, 1);
    UpdateTotalQty();
  }

  $scope.addSingleEntry = function() {
    var l = $scope.items.length;
    $scope.items.splice(l, 0, angular.copy($scope.entry));
    UpdateTotalQty();
  }

  $scope.DateChanged = function() {
    var today = new Date();
    if (today < $scope.dateValue) {
      $scope.dateValue = today;
    }
    UpdateDateDiffAsText();
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Submitting...";
    var api = "/api/rrkDailyAssemblyEmailSendApi";
    var postData = {
      "dateTimeAsUTCMilliSeconds": $scope.dateValue.getTime(),
      "items": $scope.items,
    }

    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      $scope.isLogSubmitted = true;
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
      $scope.isLogSubmitted = false;
    });
  }

  $scope.dateValue =  new Date();
  UpdateDateDiffAsText();
  $scope.entry = {modelName:"Premium Plus", quantity:5, assemblyLineName: "Line1", unit:"pc"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;
  $scope.totalQuantityProduced = 0;

}]);
