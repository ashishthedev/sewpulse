var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngGZBDailyCashController', ['$scope', '$http', function($scope, $http) {
  function FetchOpeningBalance() {
    var api = "/api/gzbDailyCashOpeningBalanceApi";
    var postData = null;
    $scope.statusNote = "Fetching opening balance ...";
    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      if (data.Initialized == true) {
        $scope.openingBalanceReadOnly = true;
        $scope.openingBalance = parseInt(data.OpeningBalance);
      } else {
        $scope.openingBalanceReadOnly = false;
      }
      UpdateUITotalAmount(); //Not necessary incidentally as of now but still doing to remain consistent with logic.
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  function FetchUnsettledAdvances() {
    //Update Unsettled Advances
    $scope.statusNote = "Fetching unsettled advances ...";
    var api = "/api/gzbDailyCashGetUnsettledAdvancesApi";
    var postData = null;
    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      $scope.unsettledAdvances = data.Items;
      if ($scope.unsettledAdvances != null) {
        for(var i=0; i<$scope.unsettledAdvances.length; i++){
          var x = $scope.unsettledAdvances[i];
          x.Amount = Math.abs(x.Amount);
          x.DateDDMMMYY = DateAsUnixTimeToDDMMMYY(x.DateAsUnixTime);
        }
      }

    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  function UpdateUITotalAmount() {
    var t = $scope.openingBalance;
    for (var i=0; i < $scope.items.length; i++) {
      t += parseInt($scope.items[i].amount);
    }
    $scope.closingBalance = t;
  }

  $scope.OpeningBalanceChanged = UpdateUITotalAmount;

  $scope.removeEntry = function(index) {
    $scope.items.splice(index, 1);
    UpdateUITotalAmount();
  }

  $scope.SettleAccountForEntry = function(index) {
    settleThisEntry = $scope.unsettledAdvances[index];
    //Update Unsettled Advances
    var api = "/api/gzbDailyCashSettleAccForOneEntryApi";
    var postData = settleThisEntry;
    $scope.statusNote = "Settling the account ...";
    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      FetchOpeningBalance();
      FetchUnsettledAdvances();
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  $scope.DateChanged = function() {
    var today = new Date();
    if (today < $scope.dateValue) {
      $scope.dateValue = today;
    }
    UpdateDateDiffAsText($scope);
  }

  $scope.addSingleCashTx = function() {
    var copyOfEntry = angular.copy($scope.entry);
    var nature = copyOfEntry.nature;
    var amount = copyOfEntry.amount;
    copyOfEntry.DateAsUnixTime = Math.floor($scope.dateValue.getTime()/1000);
    if (nature != "Received") {
      copyOfEntry.amount = -1 * Math.abs(amount);
    }
    else {
      copyOfEntry.amount = Math.abs(amount);
    }
    var l = $scope.items.length;
    $scope.items.splice(l, 0, copyOfEntry);
    UpdateUITotalAmount();
    $scope.entry.amount = 0;
  }

  $scope.submitTodaysLog = function() {
    $scope.statusNote = "Submitting...";
    var api = "/api/gzbCashBookStoreAndEmailApi";
    var postData = {
      "dateOfTransactionAsUnixTime": Math.floor($scope.dateValue.getTime()/1000),
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

  $scope.openingBalanceReadOnly = true;
  $scope.dateValue =  new Date();
  UpdateDateDiffAsText($scope);
  $scope.entry = {nature:"Spent"};
  $scope.items = [];
  $scope.statusNote = "";
  $scope.isLogSubmitted = false;
  FetchOpeningBalance();
  FetchUnsettledAdvances();

}]);
