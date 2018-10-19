$(document).ready(function() {
    var table = $('#table').DataTable( {
        lengthChange: false,
        buttons: [ 'copy', 'excel', 'pdf', 'colvis' ]
    } );
 
    table.buttons().container()
        .appendTo( '#table_wrapper .col-md-6:eq(0)' );
} );

$(document).ready(function() {
    var table1 = $('#table1').DataTable( {
        lengthChange: false,
        buttons: [ 'copy', 'excel', 'pdf', 'colvis' ]
    } );
 
    table1.buttons().container()
        .appendTo( '#table1_wrapper .col-md-6:eq(0)' );
} );

$(document).ready(function() {
    var table2 = $('#table2').DataTable( {
        lengthChange: false,
        buttons: [ 'copy', 'excel', 'pdf', 'colvis' ]
    } );
 
    table2.buttons().container()
        .appendTo( '#table2_wrapper .col-md-6:eq(0)' );
} );