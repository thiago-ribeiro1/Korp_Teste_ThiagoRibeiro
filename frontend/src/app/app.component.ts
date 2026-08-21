import { CommonModule } from '@angular/common';
import { Component, OnInit } from '@angular/core';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { HealthService, StatusServicos } from './core/services/health.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, RouterLink, RouterLinkActive, CommonModule],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss',
})
export class AppComponent implements OnInit {
  status: StatusServicos = { estoque: true, faturamento: true };

  constructor(private health: HealthService) {}

  ngOnInit(): void {
    this.verificarStatus();
    // Verificação periódica para refletir rapidamente a volta do serviço
    // de estoque após uma falha, sem exigir que o usuário recarregue a página.
    setInterval(() => this.verificarStatus(), 15000);
  }

  private verificarStatus(): void {
    this.health.verificar().subscribe((status) => (this.status = status));
  }
}
