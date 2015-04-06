var bomViewControllers = angular.module('bomViewControllers', []);

bomViewControllers.controller('ArticleDetailCtrl', ['$scope', '$routeParams', 'BOM', 'Article', function( $scope, $routeParams, BOM, Article) {
  var articleName = $routeParams.id;
  function FetchArticle() {
    $scope.statusNote = "Fetching article ...";
    $scope.article = Article.get({id:articleName},
      function(data){
        $scope.statusNote = "";
        $scope.articleFetched = true;
      },
      function(error){
        $scope.statusNote = error.status + ": " + error.data;
      });
  }

  $scope.SaveArticleOnServerAlongWithBOM = function() {
    $scope.statusNote = "Saving changes on server";
    $scope.bom.$save(function(data){
      $scope.statusNote = "";
    }, function(error){
      $scope.statusNote = error.status + ": " + error.data;
    });
  }

  $scope.deleteArticle = function() {
    $scope.statusNote = "Deleting article on server...";
    Article.remove({id:$scope.article.Name}, function(data){
      $scope.statusNote = "Article " + $scope.article.Name + " deleted...";
      $scope.articleDeleted = true;
    }, function(error){
      $scope.statusNote = error.status + ": " + error.data;
    });
  }

  $scope.allDone = function()
  {
    return $scope.articleFetched;
  }
  function InitApp(){
    $scope.articleFetched = false;

    FetchArticle();
  }
  InitApp();

}]);

bomViewControllers.controller('AdminBOMViewController', ['$scope', '$rootScope', 'BOM', function($scope, $rootScope, BOM) {
  function FetchBom() {
    $scope.statusNote = "Fetching BOM ...";
    $scope.bom = BOM.get(
      function(data){
        $scope.statusNote = "";
        var keys = [];
        for(var key in $scope.bom.AML.Articles){
          keys.push(key);
        }
        $scope.sortedArticleKeys = keys.sort();
        $scope.bomFetched = true;
      },
      function(error){
        $scope.statusNote = error.status + ": " + error.data;
      });
  }


  $scope.SaveTheChangesOnServer = function() {
    $scope.statusNote = "Saving changes on server";
    $scope.bom.$save(function(data){
      $scope.statusNote = "";
      $scope.SetGridPristine();
    }, function(error){
      $scope.statusNote = error.status + ": " + error.data;
    });
  }

  $scope.SetGridPristine = function() {
    $rootScope.isGridDirty = false;
  }
  $scope.SetGridDirty = function() {
    $rootScope.isGridDirty = true;
  }

  $scope.allDone = function()
  {
    return $scope.bomFetched;
  }

  function InitApp(){
    $scope.bomFetched = false;

    FetchBom();
    $scope.SetGridPristine();
  }
  InitApp()
}]);
