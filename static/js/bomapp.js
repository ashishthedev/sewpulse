/* App module */
var SewBOMApp = angular.module('SewBOMApp', [
    'ngRoute',
    'bomViewControllers',
    'bomServices'
    ]);

SewBOMApp.config(['$routeProvider', function($routeProvider){
  $routeProvider.
  when('/', {
    templateUrl : '/static/partials/partial_bom_view.html',
    controller: 'AdminBOMViewController'
  }).
when('/edit/article/:id', {
  templateUrl : '/static/partials/partial_article_detail.html',
  controller: 'ArticleDetailCtrl'
}).
otherwise({
  redirecTo:'/'
});

}]);


SewBOMApp.config(function($interpolateProvider) {
  $interpolateProvider.startSymbol('||');
  $interpolateProvider.endSymbol('||');
});

