var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyCashController', ['$scope', '$http', function($scope, $http) {
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


  function UpdateTotalAmount() {
    var t = $scope.openingBalance;
    for (var i=0; i < $scope.items.length; i++) {
      t += parseInt($scope.items[i].amount);
    }
    $scope.closingBalance = t;
  }

  $scope.OpeningBalanceChanged = UpdateTotalAmount;

  $scope.removeEntry = function(index) {
    $scope.items.splice(index, 1);
    UpdateTotalAmount();
  }

  $scope.DateChanged = function() {
    var today = new Date();
    if (today < $scope.dateValue) {
      $scope.dateValue = today;
    }
    UpdateDateDiffAsText();
  }

  $scope.addSingleCashTx = function() {
    var copyOfEntry = angular.copy($scope.entry);
    var nature = copyOfEntry.nature;
    var amount = copyOfEntry.amount;
    if (nature != "Received") {
      copyOfEntry.amount = -1 * Math.abs(amount);
    }
    else {
      copyOfEntry.amount = Math.abs(amount);
    }
    var l = $scope.items.length;
    $scope.items.splice(l, 0, copyOfEntry);
    UpdateTotalAmount();
    $scope.entry.amount = 0;
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Submitting...";
    var api = "/api/rrkDailyCashEmailApi";
    var postData = {
      "submissionDateTimeAsUTC": $scope.dateValue.getTime(),
      "dateOfTransactionAsUTC": $scope.dateValue.getTime(),
      "openingBalance": $scope.openingBalance,
      "closingBalance": $scope.closingBalance,
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
  $scope.entry = {nature:"Spent"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;

}]);
