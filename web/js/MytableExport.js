$(document).ready(function() {
    var table = $('#table').DataTable( {
        lengthChange: false,
        buttons: [ 'copy', 'excel', 'pdf', 'colvis' ]
    } );
 
    table.buttons().container()
        .appendTo( '#table_wrapper .col-md-6:eq(0)' );
} );
