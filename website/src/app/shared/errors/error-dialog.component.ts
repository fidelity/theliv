/*
 * Copyright FMR LLC <opensource@fidelity.com>
 *
 * SPDX-License-Identifier: Apache
 */
import { Component, Inject } from '@angular/core';
import { MAT_DIALOG_DATA } from '@angular/material/dialog';

export interface DialogData {
  status: string;
  message: string;
}

@Component({
  selector: 'error-dialog.component',
  templateUrl: 'error-dialog.component.html',
  styleUrls: ['error-dialog.component.scss']
})
export class ErrorDialog {

  constructor(@Inject(MAT_DIALOG_DATA) public data: DialogData) {}


}


