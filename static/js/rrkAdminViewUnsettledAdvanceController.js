var appMod = angular.module('ngSEWPulseApp', []);

appMod.controller('ngRRKAdminUnsettledAdvanceController', ['$scope', '$http', function($scope, $http) {

  function FetchCashInHand() {
    $scope.statusNote = "Fetching cash in hand ...";
    var api = "/api/rrkDailyCashOpeningBalanceApi";
    var postData = null;

    $http.post(api, postData).success(function(data, status, headers, config) {
      $scope.statusNote = "";
      if (data.Initialized == true) {
        $scope.CashInHand = parseInt(data.OpeningBalance);
      } else {
        $scope.CashInHand = 0;
      }
    }).error(function(data, status, headers, config){
      $scope.statusNote = status + ": " + data;
    });
  }

  function FetchUnsettledAdvances() {
    var api = "/api/rrkDailyCashGetUnsettledAdvancesApi";
    var postData = null;
    $scope.statusNote = "Fetching unsettled advances ...";
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

  FetchUnsettledAdvances();
  FetchCashInHand();
}]);
