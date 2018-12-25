/* When the user scrolls down, hide the navbar. When the user scrolls up, show the navbar */
var prevScrollpos = window.pageYOffset;
window.onscroll = function() {
  var currentScrollPos = window.pageYOffset;
  if (prevScrollpos < currentScrollPos) {
    document.getElementById("navbar").style.top = "-90px";
  } else {
    document.getElementById("navbar").style.top = "0";
  }
  prevScrollpos = currentScrollPos;
}

$(document).ready(function() {
    var table = $('#table').DataTable( 
        {
//           "scrollY":400,
//           "scrollX":true,
//           "scroller":false

        "scrollX":true,
        "scroller":false,
        lengthChange: false,
        buttons: [ {extend: 'copy',  filename: 'Data export'}, 
                   {extend: 'csv',  filename: 'Data export'}, 
                   {extend: 'excel',  filename: 'Data export'}, 
                   {extend: 'pdf',  filename: 'Data export'},
                   {extend: 'print',  filename: 'Data export'},
                   'colvis']
    } );
 
    table.buttons().container()
        .appendTo( '#table_wrapper .col-md-6:eq(0)' );
} );

$(document).ready(function() {
    var table1 = $('#table1').DataTable( {
        "scrollX":true,
        "scroller":false,
        lengthChange: false,
        buttons: [ {extend: 'copy',  filename: 'Data export'}, 
                   {extend: 'csv',  filename: 'Data export'}, 
                   {extend: 'excel',  filename: 'Data export'}, 
                   {extend: 'pdf',  filename: 'Data export'},
                   {extend: 'print',  filename: 'Data export'},
                   'colvis']
    } );
 
    table1.buttons().container()
        .appendTo( '#table1_wrapper .col-md-6:eq(0)' );
} );

$(document).ready(function() {
    var table2 = $('#table2').DataTable( {
        "scrollX":true,
        "scroller":false,
        lengthChange: false,
        buttons: [ {extend: 'copy',  filename: 'Data export'}, 
                   {extend: 'csv',  filename: 'Data export'}, 
                   {extend: 'excel',  filename: 'Data export'}, 
                   {extend: 'pdf',  filename: 'Data export'},
                   {extend: 'print',  filename: 'Data export'},
                   'colvis']
    } );
 
    table2.buttons().container()
        .appendTo( '#table2_wrapper .col-md-6:eq(0)' );
} );

