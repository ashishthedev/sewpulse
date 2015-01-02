var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKDailyCashController', ['$scope', '$http', function($scope, $http) {

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
  $scope.entry = {nature:"Spent"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;

}]);
